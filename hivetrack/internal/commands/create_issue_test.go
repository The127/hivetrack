package commands_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

func setupProject(t *testing.T, db *inmemory.DbContext, actor *models.User) *models.Project {
	t.Helper()
	project := models.NewProject(actor.GetId(), "myproject", "My Project", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))
	return project
}

func TestHandleCreateIssue_QuickCapture(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := setupProject(t, db, actor)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	// Quick-capture: title only → triaged=false
	result, err := commands.HandleCreateIssue(ctx, commands.CreateIssueCommand{
		ProjectSlug: project.GetSlug(),
		Title:       "Fix the bug",
		Type:        models.IssueTypeTask,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, result.Number, 0)

	issue, err := db.Issues().GetByNumber(context.Background(), project.GetId(), result.Number)
	require.NoError(t, err)
	require.NotNil(t, issue)
	assert.Equal(t, "Fix the bug", issue.GetTitle())
	assert.False(t, issue.GetTriaged(), "quick-capture should be untriaged")
	assert.Equal(t, actor.GetId(), *issue.GetReporterID())
}

func TestHandleCreateIssue_WithStatus_IsTriaged(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := setupProject(t, db, actor)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	status := models.IssueStatusTodo
	result, err := commands.HandleCreateIssue(ctx, commands.CreateIssueCommand{
		ProjectSlug: project.GetSlug(),
		Title:       "Planned task",
		Type:        models.IssueTypeTask,
		Status:      &status,
	})
	require.NoError(t, err)

	issue, err := db.Issues().GetByNumber(context.Background(), project.GetId(), result.Number)
	require.NoError(t, err)
	assert.True(t, issue.GetTriaged(), "issue with explicit status should be triaged")
}

func TestHandleCreateIssue_WithAssignees(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	other := models.NewUser("sub2", "other@example.com", "other@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	require.NoError(t, db.Users().Upsert(context.Background(), other))
	project := setupProject(t, db, actor)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	result, err := commands.HandleCreateIssue(ctx, commands.CreateIssueCommand{
		ProjectSlug: project.GetSlug(),
		Title:       "Assigned task",
		Type:        models.IssueTypeTask,
		AssigneeIDs: []uuid.UUID{actor.GetId(), other.GetId()},
	})
	require.NoError(t, err)

	issue, err := db.Issues().GetByNumber(context.Background(), project.GetId(), result.Number)
	require.NoError(t, err)
	assert.Len(t, issue.GetAssignees(), 2)
}

func TestHandleCreateIssue_WithParent(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := setupProject(t, db, actor)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	// Create epic via handler so the issue counter advances
	epicResult, err := commands.HandleCreateIssue(ctx, commands.CreateIssueCommand{
		ProjectSlug: project.GetSlug(),
		Title:       "My Epic",
		Type:        models.IssueTypeEpic,
	})
	require.NoError(t, err)

	result, err := commands.HandleCreateIssue(ctx, commands.CreateIssueCommand{
		ProjectSlug: project.GetSlug(),
		Title:       "Child task",
		Type:        models.IssueTypeTask,
		ParentID:    &epicResult.ID,
	})
	require.NoError(t, err)

	issue, err := db.Issues().GetByNumber(context.Background(), project.GetId(), result.Number)
	require.NoError(t, err)
	require.NotNil(t, issue.GetParentID())
	assert.Equal(t, epicResult.ID, *issue.GetParentID())
}

func TestHandleCreateIssue_EpicCannotHaveParent(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := setupProject(t, db, actor)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	epicResult, err := commands.HandleCreateIssue(ctx, commands.CreateIssueCommand{
		ProjectSlug: project.GetSlug(),
		Title:       "My Epic",
		Type:        models.IssueTypeEpic,
	})
	require.NoError(t, err)

	_, err = commands.HandleCreateIssue(ctx, commands.CreateIssueCommand{
		ProjectSlug: project.GetSlug(),
		Title:       "Another epic",
		Type:        models.IssueTypeEpic,
		ParentID:    &epicResult.ID,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, models.ErrBadRequest)
}

func TestHandleCreateIssue_ProjectNotFound(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	_, err := commands.HandleCreateIssue(ctx, commands.CreateIssueCommand{
		ProjectSlug: "nonexistent",
		Title:       "Test",
		Type:        models.IssueTypeTask,
	})
	require.Error(t, err)
}
