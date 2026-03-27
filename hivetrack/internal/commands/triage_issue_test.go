package commands_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

func TestHandleTriageIssue_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	reporterID := actor.GetId()
	untriaged := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "Incoming",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, false, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(untriaged)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleTriageIssue(ctx, commands.TriageIssueCommand{
		IssueID: untriaged.GetId(),
		Status:  models.IssueStatusInProgress,
	})
	require.NoError(t, err)

	issue, err := db.Issues().GetByID(context.Background(), untriaged.GetId())
	require.NoError(t, err)
	assert.True(t, issue.GetTriaged())
	assert.Equal(t, models.IssueStatusInProgress, issue.GetStatus())
}

func TestHandleTriageIssue_WithPriorityAndEstimate(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	reporterID := actor.GetId()
	untriaged := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "Incoming",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, false, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(untriaged)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	priority := models.IssuePriorityHigh
	estimate := models.IssueEstimateM
	_, err := commands.HandleTriageIssue(ctx, commands.TriageIssueCommand{
		IssueID:  untriaged.GetId(),
		Status:   models.IssueStatusTodo,
		Priority: &priority,
		Estimate: &estimate,
	})
	require.NoError(t, err)

	issue, err := db.Issues().GetByID(context.Background(), untriaged.GetId())
	require.NoError(t, err)
	assert.True(t, issue.GetTriaged())
	assert.Equal(t, models.IssuePriorityHigh, issue.GetPriority())
	assert.Equal(t, models.IssueEstimateM, issue.GetEstimate())
}

func TestHandleTriageIssue_WithoutPriorityAndEstimate_LeavesDefaults(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	reporterID := actor.GetId()
	untriaged := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "Incoming",
		models.IssueStatusTodo, models.IssuePriorityLow, models.IssueEstimateS,
		&reporterID, false, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(untriaged)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleTriageIssue(ctx, commands.TriageIssueCommand{
		IssueID: untriaged.GetId(),
		Status:  models.IssueStatusInProgress,
	})
	require.NoError(t, err)

	issue, err := db.Issues().GetByID(context.Background(), untriaged.GetId())
	require.NoError(t, err)
	assert.Equal(t, models.IssuePriorityLow, issue.GetPriority())
	assert.Equal(t, models.IssueEstimateS, issue.GetEstimate())
}
