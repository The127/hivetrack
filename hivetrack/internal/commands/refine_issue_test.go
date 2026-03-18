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

func TestHandleRefineIssue_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	reporterID := actor.GetId()
	unrefined := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "Some task",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(unrefined)
	require.NoError(t, db.SaveChanges(context.Background()))

	assert.False(t, unrefined.GetRefined())

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleRefineIssue(ctx, commands.RefineIssueCommand{
		IssueID: unrefined.GetId(),
	})
	require.NoError(t, err)

	issue, err := db.Issues().GetByID(context.Background(), unrefined.GetId())
	require.NoError(t, err)
	assert.True(t, issue.GetRefined())
}

func TestHandleRefineIssue_NotFound(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleRefineIssue(ctx, commands.RefineIssueCommand{
		IssueID: models.NewBaseModel().GetId(),
	})
	require.ErrorIs(t, err, models.ErrNotFound)
}
