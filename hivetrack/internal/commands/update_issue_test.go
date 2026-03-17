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

func newTestIssue(projectID uuid.UUID, reporterID uuid.UUID, number int) *models.Issue {
	return models.NewIssue(
		projectID, number, models.IssueTypeTask, "Original title",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
}

func TestHandleUpdateIssue_ChangeTitle(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	issue := newTestIssue(project.GetId(), actor.GetId(), 1)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	newTitle := "New title"
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID: issue.GetId(),
		Title:   &newTitle,
	})
	require.NoError(t, err)

	updated, err := db.Issues().GetByID(context.Background(), issue.GetId())
	require.NoError(t, err)
	assert.Equal(t, "New title", updated.GetTitle())
}

func TestHandleUpdateIssue_SetOnHold(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	issue := newTestIssue(project.GetId(), actor.GetId(), 1)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	onHold := true
	reason := models.HoldReasonWaitingOnCustomer
	note := "waiting for response"
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID:    issue.GetId(),
		OnHold:     &onHold,
		HoldReason: &reason,
		HoldNote:   &note,
	})
	require.NoError(t, err)

	updated, err := db.Issues().GetByID(context.Background(), issue.GetId())
	require.NoError(t, err)
	assert.True(t, updated.GetOnHold())
	require.NotNil(t, updated.GetHoldReason())
	assert.Equal(t, models.HoldReasonWaitingOnCustomer, *updated.GetHoldReason())
}
