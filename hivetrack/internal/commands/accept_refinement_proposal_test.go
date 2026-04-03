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

func TestAcceptRefinementProposal_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	reporterID := actor.GetId()
	issue := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "Old title",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	// Create session with a proposal message
	session := models.NewRefinementSession(issue.GetId())
	require.NoError(t, db.Refinements().CreateSession(context.Background(), session))

	proposalMsg := models.NewRefinementMessage(session.ID, models.RefinementRoleAssistant,
		"Here is my proposal", models.RefinementMessageTypeProposal,
		&models.RefinementProposal{Title: "New title", Description: "New description"})
	require.NoError(t, db.Refinements().AddMessage(context.Background(), proposalMsg))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleAcceptRefinementProposal(ctx, commands.AcceptRefinementProposalCommand{
		IssueID: issue.GetId(),
	})
	require.NoError(t, err)

	// Issue was updated
	updated, err := db.Issues().GetByID(context.Background(), issue.GetId())
	require.NoError(t, err)
	assert.Equal(t, "New title", updated.GetTitle())
	assert.Equal(t, "New description", *updated.GetDescription())
	assert.True(t, updated.GetRefined())

	// Session was completed
	active, err := db.Refinements().GetActiveSession(context.Background(), issue.GetId())
	require.NoError(t, err)
	assert.Nil(t, active)
}

func TestAcceptRefinementProposal_NoActiveSession(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	reporterID := actor.GetId()
	issue := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "Task",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleAcceptRefinementProposal(ctx, commands.AcceptRefinementProposalCommand{
		IssueID: issue.GetId(),
	})
	require.ErrorIs(t, err, models.ErrNotFound)
}

func TestAcceptRefinementProposal_NoProposalMessage(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	reporterID := actor.GetId()
	issue := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "Task",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	// Session exists but only has regular messages, no proposal
	session := models.NewRefinementSession(issue.GetId())
	require.NoError(t, db.Refinements().CreateSession(context.Background(), session))
	msg := models.NewRefinementMessage(session.ID, models.RefinementRoleAssistant, "Just a question", models.RefinementMessageTypeMessage, nil)
	require.NoError(t, db.Refinements().AddMessage(context.Background(), msg))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleAcceptRefinementProposal(ctx, commands.AcceptRefinementProposalCommand{
		IssueID: issue.GetId(),
	})
	require.ErrorIs(t, err, models.ErrNotFound)
}
