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

func setupProject(t *testing.T, db *inmemory.DbContext, actor models.User) *models.Project {
	t.Helper()
	project := &models.Project{
		ID:        uuid.New(),
		Slug:      "myproject",
		Name:      "My Project",
		Archetype: models.ProjectArchetypeSoftware,
		CreatedBy: actor.ID,
	}
	require.NoError(t, db.Projects().Insert(context.Background(), project))
	return project
}

func TestHandleCreateIssue_QuickCapture(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))
	project := setupProject(t, db, actor)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	// Quick-capture: title only → triaged=false
	result, err := commands.HandleCreateIssue(ctx, commands.CreateIssueCommand{
		ProjectSlug: project.Slug,
		Title:       "Fix the bug",
		Type:        models.IssueTypeTask,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, result.Number, 0)

	issue, err := db.Issues().GetByNumber(context.Background(), project.ID, result.Number)
	require.NoError(t, err)
	require.NotNil(t, issue)
	assert.Equal(t, "Fix the bug", issue.Title)
	assert.False(t, issue.Triaged, "quick-capture should be untriaged")
	assert.Equal(t, actor.ID, *issue.ReporterID)
}

func TestHandleCreateIssue_WithStatus_IsTriaged(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))
	project := setupProject(t, db, actor)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	status := models.IssueStatusTodo
	result, err := commands.HandleCreateIssue(ctx, commands.CreateIssueCommand{
		ProjectSlug: project.Slug,
		Title:       "Planned task",
		Type:        models.IssueTypeTask,
		Status:      &status,
	})
	require.NoError(t, err)

	issue, err := db.Issues().GetByNumber(context.Background(), project.ID, result.Number)
	require.NoError(t, err)
	assert.True(t, issue.Triaged, "issue with explicit status should be triaged")
}

func TestHandleCreateIssue_WithAssignees(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	other := models.User{ID: uuid.New(), Sub: "sub2", Email: "other@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))
	require.NoError(t, db.Users().Upsert(context.Background(), &other))
	project := setupProject(t, db, actor)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	result, err := commands.HandleCreateIssue(ctx, commands.CreateIssueCommand{
		ProjectSlug: project.Slug,
		Title:       "Assigned task",
		Type:        models.IssueTypeTask,
		AssigneeIDs: []uuid.UUID{actor.ID, other.ID},
	})
	require.NoError(t, err)

	issue, err := db.Issues().GetByNumber(context.Background(), project.ID, result.Number)
	require.NoError(t, err)
	assert.Len(t, issue.Assignees, 2)
}

func TestHandleCreateIssue_ProjectNotFound(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	_, err := commands.HandleCreateIssue(ctx, commands.CreateIssueCommand{
		ProjectSlug: "nonexistent",
		Title:       "Test",
		Type:        models.IssueTypeTask,
	})
	require.Error(t, err)
}
