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

func TestCreateIssueLink_BlocksAutoSetsHoldOnTarget(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	reporterID := actor.GetId()
	blocker := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "Blocker",
		models.IssueStatusInProgress, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	blocked := models.NewIssue(
		project.GetId(), 2, models.IssueTypeTask, "Blocked",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(blocker)
	db.Issues().Insert(blocked)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleCreateIssueLink(ctx, commands.CreateIssueLinkCommand{
		SourceIssueID: blocker.GetId(),
		TargetIssueID: blocked.GetId(),
		LinkType:      models.LinkTypeBlocks,
	})
	require.NoError(t, err)

	updated, err := db.Issues().GetByID(context.Background(), blocked.GetId())
	require.NoError(t, err)
	assert.True(t, updated.GetOnHold())
	assert.Equal(t, models.HoldReasonBlockedByIssue, *updated.GetHoldReason())
}

func TestCreateIssueLink_AlreadyOnHoldDoesNotOverwrite(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	reporterID := actor.GetId()
	blocker := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "Blocker",
		models.IssueStatusInProgress, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	blocked := models.NewIssue(
		project.GetId(), 2, models.IssueTypeTask, "Blocked",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(blocker)
	db.Issues().Insert(blocked)
	require.NoError(t, db.SaveChanges(context.Background()))

	// Set on hold with a different reason first
	reason := models.HoldReasonWaitingOnCustomer
	now := blocked.GetUpdatedAt()
	note := "waiting for customer response"
	blocked.SetHold(true, &reason, &now, &note)
	db.Issues().Update(blocked)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleCreateIssueLink(ctx, commands.CreateIssueLinkCommand{
		SourceIssueID: blocker.GetId(),
		TargetIssueID: blocked.GetId(),
		LinkType:      models.LinkTypeBlocks,
	})
	require.NoError(t, err)

	// Should keep the original hold reason
	updated, err := db.Issues().GetByID(context.Background(), blocked.GetId())
	require.NoError(t, err)
	assert.True(t, updated.GetOnHold())
	assert.Equal(t, models.HoldReasonWaitingOnCustomer, *updated.GetHoldReason())
}
