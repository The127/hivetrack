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
