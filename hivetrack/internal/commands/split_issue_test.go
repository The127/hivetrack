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

func setupTask(ctx context.Context, t *testing.T, db *inmemory.DbContext, project *models.Project) *models.Issue {
	t.Helper()
	priority := models.IssuePriorityMedium
	estimate := models.IssueEstimateM
	result, err := commands.HandleCreateIssue(ctx, commands.CreateIssueCommand{
		ProjectSlug: project.GetSlug(),
		Title:       "Original task",
		Type:        models.IssueTypeTask,
		Priority:    &priority,
		Estimate:    &estimate,
	})
	require.NoError(t, err)
	task, err := db.Issues().GetByNumber(ctx, project.GetId(), result.Number)
	require.NoError(t, err)
	require.NotNil(t, task)
	return task
}

func TestHandleSplitIssue_HappyPath(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := setupProject(t, db, actor)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	task := setupTask(ctx, t, db, project)

	result, err := commands.HandleSplitIssue(ctx, commands.SplitIssueCommand{
		IssueID:   task.GetId(),
		NewTitles: []string{"Part A", "Part B"},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.NewIssues, 2)

	// Original issue should be cancelled
	original, err := db.Issues().GetByID(context.Background(), task.GetId())
	require.NoError(t, err)
	assert.Equal(t, models.IssueStatusCancelled, original.GetStatus())

	// New issues should exist with correct titles
	issue1, err := db.Issues().GetByNumber(context.Background(), project.GetId(), result.NewIssues[0].Number)
	require.NoError(t, err)
	require.NotNil(t, issue1)
	assert.Equal(t, "Part A", issue1.GetTitle())
	assert.True(t, issue1.GetTriaged())
	assert.Equal(t, models.IssueStatusTodo, issue1.GetStatus())
	assert.Equal(t, models.IssuePriorityMedium, issue1.GetPriority())
	assert.Equal(t, models.IssueEstimateM, issue1.GetEstimate())

	issue2, err := db.Issues().GetByNumber(context.Background(), project.GetId(), result.NewIssues[1].Number)
	require.NoError(t, err)
	require.NotNil(t, issue2)
	assert.Equal(t, "Part B", issue2.GetTitle())

	// Links should be created
	links, err := db.Issues().ListLinks(context.Background(), task.GetId())
	require.NoError(t, err)
	assert.Len(t, links, 2)
	for _, l := range links {
		assert.Equal(t, task.GetId(), l.SourceIssueID)
		assert.Equal(t, models.LinkTypeRelatesTo, l.LinkType)
	}
}

func TestHandleSplitIssue_MinTitlesGuard(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := setupProject(t, db, actor)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	task := setupTask(ctx, t, db, project)

	_, err := commands.HandleSplitIssue(ctx, commands.SplitIssueCommand{
		IssueID:   task.GetId(),
		NewTitles: []string{"Only one title"},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, models.ErrBadRequest)
}

func TestHandleSplitIssue_TerminalIssueGuard(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := setupProject(t, db, actor)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	// Create a task then mark it done directly
	task := setupTask(ctx, t, db, project)
	task.SetStatus(models.IssueStatusDone)
	db.Issues().Update(task)
	require.NoError(t, db.SaveChanges(context.Background()))

	_, err := commands.HandleSplitIssue(ctx, commands.SplitIssueCommand{
		IssueID:   task.GetId(),
		NewTitles: []string{"Part A", "Part B"},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, models.ErrBadRequest)
}

func TestHandleSplitIssue_EpicGuard(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := setupProject(t, db, actor)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	epicResult, err := commands.HandleCreateIssue(ctx, commands.CreateIssueCommand{
		ProjectSlug: project.GetSlug(),
		Title:       "An epic",
		Type:        models.IssueTypeEpic,
	})
	require.NoError(t, err)

	_, err = commands.HandleSplitIssue(ctx, commands.SplitIssueCommand{
		IssueID:   epicResult.ID,
		NewTitles: []string{"Part A", "Part B"},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, models.ErrBadRequest)
}

func TestHandleSplitIssue_InheritsProperties(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	assignee := models.NewUser("sub2", "assignee@example.com", "assignee@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	require.NoError(t, db.Users().Upsert(context.Background(), assignee))
	project := setupProject(t, db, actor)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	priority := models.IssuePriorityHigh
	estimate := models.IssueEstimateL
	taskResult, err := commands.HandleCreateIssue(ctx, commands.CreateIssueCommand{
		ProjectSlug: project.GetSlug(),
		Title:       "Task with properties",
		Type:        models.IssueTypeTask,
		Priority:    &priority,
		Estimate:    &estimate,
		AssigneeIDs: []uuid.UUID{assignee.GetId()},
	})
	require.NoError(t, err)

	result, err := commands.HandleSplitIssue(ctx, commands.SplitIssueCommand{
		IssueID:   taskResult.ID,
		NewTitles: []string{"Child 1", "Child 2"},
	})
	require.NoError(t, err)

	for _, r := range result.NewIssues {
		ni, err := db.Issues().GetByNumber(context.Background(), project.GetId(), r.Number)
		require.NoError(t, err)
		assert.Equal(t, models.IssuePriorityHigh, ni.GetPriority())
		assert.Equal(t, models.IssueEstimateL, ni.GetEstimate())
		assert.Equal(t, []uuid.UUID{assignee.GetId()}, ni.GetAssignees())
	}
}
