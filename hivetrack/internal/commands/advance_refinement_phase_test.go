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

func setupAdvancePhaseTest(t *testing.T) (*inmemory.DbContext, *models.Issue, *models.RefinementSession, *spyPublisher) {
	t.Helper()
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "proj", "Proj", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	reporterID := actor.GetId()
	desc := "A description"
	issue := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "Task",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		&desc, nil, nil, nil, nil,
	)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	session := models.NewRefinementSession(issue.GetId())
	require.NoError(t, db.Refinements().CreateSession(context.Background(), session))

	pub := &spyPublisher{}
	return db, issue, session, pub
}

func TestAdvanceRefinementPhase_NextPhase(t *testing.T) {
	db, issue, _, pub := setupAdvancePhaseTest(t)
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")

	handler := commands.NewAdvanceRefinementPhaseHandler(pub, func(uuid.UUID) {})
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	result, err := handler(ctx, commands.AdvanceRefinementPhaseCommand{
		IssueID: issue.GetId(),
	})
	require.NoError(t, err)
	assert.Equal(t, "main_scenario", result.Phase)

	// Session phase was updated
	session, err := db.Refinements().GetActiveSession(context.Background(), issue.GetId())
	require.NoError(t, err)
	assert.Equal(t, models.RefinementPhaseMainScenario, session.CurrentPhase)

	// NATS request was published with new phase
	require.Len(t, pub.published, 1)
	assert.Equal(t, "main_scenario", pub.published[0].Phase)
}

func TestAdvanceRefinementPhase_AllPhasesSequentially(t *testing.T) {
	db, issue, _, pub := setupAdvancePhaseTest(t)
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")

	handler := commands.NewAdvanceRefinementPhaseHandler(pub, func(uuid.UUID) {})
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	expected := []string{"main_scenario", "extensions", "acceptance_criteria"}
	for _, want := range expected {
		result, err := handler(ctx, commands.AdvanceRefinementPhaseCommand{
			IssueID: issue.GetId(),
		})
		require.NoError(t, err)
		assert.Equal(t, want, result.Phase)
	}
}

func TestAdvanceRefinementPhase_AlreadyAtLastPhase(t *testing.T) {
	db, issue, session, pub := setupAdvancePhaseTest(t)
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")

	// Set to last phase
	require.NoError(t, db.Refinements().UpdateSessionPhase(context.Background(), session.ID, models.RefinementPhaseAcceptanceCriteria))

	handler := commands.NewAdvanceRefinementPhaseHandler(pub, func(uuid.UUID) {})
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	_, err := handler(ctx, commands.AdvanceRefinementPhaseCommand{
		IssueID: issue.GetId(),
	})
	require.ErrorIs(t, err, models.ErrBadRequest)
	assert.Empty(t, pub.published)
}

func TestAdvanceRefinementPhase_Regression(t *testing.T) {
	db, issue, session, pub := setupAdvancePhaseTest(t)
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")

	// Advance to extensions
	require.NoError(t, db.Refinements().UpdateSessionPhase(context.Background(), session.ID, models.RefinementPhaseExtensions))

	handler := commands.NewAdvanceRefinementPhaseHandler(pub, func(uuid.UUID) {})
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	result, err := handler(ctx, commands.AdvanceRefinementPhaseCommand{
		IssueID:     issue.GetId(),
		TargetPhase: "actor_goal",
	})
	require.NoError(t, err)
	assert.Equal(t, "actor_goal", result.Phase)

	// Session phase was updated
	s, err := db.Refinements().GetActiveSession(context.Background(), issue.GetId())
	require.NoError(t, err)
	assert.Equal(t, models.RefinementPhaseActorGoal, s.CurrentPhase)
}

func TestAdvanceRefinementPhase_InvalidTargetPhase(t *testing.T) {
	db, issue, _, pub := setupAdvancePhaseTest(t)
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")

	handler := commands.NewAdvanceRefinementPhaseHandler(pub, func(uuid.UUID) {})
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	_, err := handler(ctx, commands.AdvanceRefinementPhaseCommand{
		IssueID:     issue.GetId(),
		TargetPhase: "nonsense",
	})
	require.ErrorIs(t, err, models.ErrBadRequest)
	assert.Empty(t, pub.published)
}

func TestAdvanceRefinementPhase_NoActiveSession(t *testing.T) {
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

	pub := &spyPublisher{}
	handler := commands.NewAdvanceRefinementPhaseHandler(pub, func(uuid.UUID) {})
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	_, err := handler(ctx, commands.AdvanceRefinementPhaseCommand{
		IssueID: issue.GetId(),
	})
	require.ErrorIs(t, err, models.ErrNotFound)
}

func TestAdvanceRefinementPhase_PublishesCorrectPhase(t *testing.T) {
	db, issue, _, pub := setupAdvancePhaseTest(t)
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")

	// Add a message so we can verify history is included
	msg := models.NewRefinementMessage(
		issue.GetId(), // will use session ID below
		models.RefinementRoleAssistant, "Who is the actor?",
		models.RefinementMessageTypeMessage, models.RefinementPhaseActorGoal, nil,
	)
	// Fix: get session to use its ID
	session, err := db.Refinements().GetActiveSession(context.Background(), issue.GetId())
	require.NoError(t, err)
	msg.SessionID = session.ID
	require.NoError(t, db.Refinements().AddMessage(context.Background(), msg))

	handler := commands.NewAdvanceRefinementPhaseHandler(pub, func(uuid.UUID) {})
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	result, err := handler(ctx, commands.AdvanceRefinementPhaseCommand{
		IssueID: issue.GetId(),
	})
	require.NoError(t, err)
	assert.Equal(t, "main_scenario", result.Phase)

	require.Len(t, pub.published, 1)
	req := pub.published[0]
	assert.Equal(t, "main_scenario", req.Phase)
	assert.Equal(t, session.ID, req.SessionID)
	assert.Equal(t, "proj", req.ProjectSlug)
	require.Len(t, req.Messages, 1)
	assert.Equal(t, "assistant", req.Messages[0].Role)
}
