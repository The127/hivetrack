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

func TestHandleBatchUpdateIssues_UpdatesMultipleIssues(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	reporterID := actor.GetId()
	issue1 := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "Issue 1",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	issue2 := models.NewIssue(
		project.GetId(), 2, models.IssueTypeTask, "Issue 2",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(issue1)
	db.Issues().Insert(issue2)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	priority := models.IssuePriorityHigh
	status := models.IssueStatusInProgress
	result, err := commands.HandleBatchUpdateIssues(ctx, commands.BatchUpdateIssuesCommand{
		ProjectID:    project.GetId(),
		IssueNumbers: []int{1, 2},
		Priority:     &priority,
		Status:       &status,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, result.Updated)

	updated1, _ := db.Issues().GetByID(context.Background(), issue1.GetId())
	assert.Equal(t, models.IssuePriorityHigh, updated1.GetPriority())
	assert.Equal(t, models.IssueStatusInProgress, updated1.GetStatus())

	updated2, _ := db.Issues().GetByID(context.Background(), issue2.GetId())
	assert.Equal(t, models.IssuePriorityHigh, updated2.GetPriority())
	assert.Equal(t, models.IssueStatusInProgress, updated2.GetStatus())
}

func TestHandleBatchUpdateIssues_ClearSprintID(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	sprintID := uuid.New()
	reporterID := actor.GetId()
	issue := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "In Sprint",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, &sprintID, nil, nil, nil,
	)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleBatchUpdateIssues(ctx, commands.BatchUpdateIssuesCommand{
		ProjectID:     project.GetId(),
		IssueNumbers:  []int{1},
		ClearSprintID: true,
	})
	require.NoError(t, err)

	updated, _ := db.Issues().GetByID(context.Background(), issue.GetId())
	assert.Nil(t, updated.GetSprintID())
}

func TestHandleBatchUpdateIssues_UnknownIssueReturnsError(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	reporterID := actor.GetId()
	issue1 := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "Issue 1",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(issue1)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	priority := models.IssuePriorityHigh
	_, err := commands.HandleBatchUpdateIssues(ctx, commands.BatchUpdateIssuesCommand{
		ProjectID:    project.GetId(),
		IssueNumbers: []int{1, 999},
		Priority:     &priority,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, models.ErrNotFound)
}
