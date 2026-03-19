package events_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/the127/hivetrack/internal/events"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

// stubScenarioGenerator returns a fixed GIVEN/WHEN/THEN text for any description.
type stubScenarioGenerator struct {
	scenarios string
}

func (s *stubScenarioGenerator) GenerateScenarios(_ context.Context, _ string) (string, error) {
	return s.scenarios, nil
}

// vagueScenarioGenerator simulates a generator that rejects vague criteria.
type vagueScenarioGenerator struct{}

func (v *vagueScenarioGenerator) GenerateScenarios(_ context.Context, _ string) (string, error) {
	return "", events.ErrVagueCriteria
}

// spyScenarioGenerator records whether GenerateScenarios was called.
type spyScenarioGenerator struct {
	called bool
}

func (s *spyScenarioGenerator) GenerateScenarios(_ context.Context, _ string) (string, error) {
	s.called = true
	return "", nil
}

const cannedScenarios = `GIVEN a user is logged in
WHEN they click submit
THEN the form is saved`

func newRefinedEpicWithDescription(projectID uuid.UUID, description string) *models.Issue {
	reporterID := uuid.New()
	desc := description
	issue := models.NewIssue(
		projectID, 2, models.IssueTypeEpic, "Big epic feature",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		&desc, nil, nil, nil, nil,
	)
	issue.SetRefined(true)
	return issue
}

func newRefinedTaskWithDescription(projectID uuid.UUID, description string) *models.Issue {
	reporterID := uuid.New()
	desc := description
	issue := models.NewIssue(
		projectID, 1, models.IssueTypeTask, "Add login feature",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		&desc, nil, nil, nil, nil,
	)
	issue.SetRefined(true)
	return issue
}

func TestHivemind_PostsScenariosComment_WhenRefinedTaskHasAcceptanceCriteria(t *testing.T) {
	db := inmemory.NewDbContext()
	ctx := testutil.ContextWithDb(db)

	projectID := uuid.New()
	issue := newRefinedTaskWithDescription(projectID, "Users must be able to log in with email and password. On success redirect to dashboard.")
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	gen := &stubScenarioGenerator{scenarios: cannedScenarios}
	handler := events.HandleIssueRefinedForHivemind(gen)

	err := handler(ctx, events.IssueRefinedPayload{
		IssueID: issue.GetId(),
		ActorID: uuid.New(),
	})
	require.NoError(t, err)

	// A comment must have been created on the issue.
	comments, total, err := db.Comments().List(context.Background(), issue.GetId(), 0, 0)
	require.NoError(t, err)
	require.Equal(t, 1, total, "expected exactly one comment to be posted by Hivemind")
	require.Len(t, comments, 1)

	comment := comments[0]

	// The comment body must contain the generated scenarios.
	assert.Contains(t, comment.GetBody(), cannedScenarios)

	// The comment must be attributed to Hivemind (nil authorID, fixed email and name).
	assert.Nil(t, comment.GetAuthorID(), "Hivemind comment must have no internal authorID")
	require.NotNil(t, comment.GetAuthorEmail())
	assert.Equal(t, "hivemind@hivetrack.internal", *comment.GetAuthorEmail())
	require.NotNil(t, comment.GetAuthorName())
	assert.True(t, strings.EqualFold("hivemind", *comment.GetAuthorName()), "author name should identify Hivemind")
}

type refinedEpicHandlerSetup struct {
	db      *inmemory.DbContext
	ctx     context.Context
	issue   *models.Issue
	spy     *spyScenarioGenerator
	handler func(context.Context, events.IssueRefinedPayload) error
}

func setupRefinedEpicHandler(t *testing.T) refinedEpicHandlerSetup {
	t.Helper()
	db := inmemory.NewDbContext()
	ctx := testutil.ContextWithDb(db)

	projectID := uuid.New()
	issue := newRefinedEpicWithDescription(projectID, "This epic covers the entire authentication surface area.")
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	spy := &spyScenarioGenerator{}
	handler := events.HandleIssueRefinedForHivemind(spy)

	return refinedEpicHandlerSetup{
		db:      db,
		ctx:     ctx,
		issue:   issue,
		spy:     spy,
		handler: handler,
	}
}

func TestHivemind_SkipsEpicIssue_WhenRefinedEventReceived(t *testing.T) {
	s := setupRefinedEpicHandler(t)

	err := s.handler(s.ctx, events.IssueRefinedPayload{
		IssueID: s.issue.GetId(),
		ActorID: uuid.New(),
	})
	require.NoError(t, err)

	assert.False(t, s.spy.called, "GenerateScenarios must not be called for epic-type issues")

	_, total, err := s.db.Comments().List(context.Background(), s.issue.GetId(), 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, total, "no comment must be posted for epic-type issues")
}

func TestHivemind_RevertsRefinedFlag_WhenCriteriaAreInsufficient(t *testing.T) {
	db := inmemory.NewDbContext()
	ctx := testutil.ContextWithDb(db)

	projectID := uuid.New()
	issue := newRefinedTaskWithDescription(projectID, "It should work better.")
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	gen := &vagueScenarioGenerator{}
	handler := events.HandleIssueRefinedForHivemind(gen)

	err := handler(ctx, events.IssueRefinedPayload{
		IssueID: issue.GetId(),
		ActorID: uuid.New(),
	})
	require.NoError(t, err)

	// The refined flag must have been reverted to false.
	updated, err := db.Issues().GetByID(context.Background(), issue.GetId())
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.False(t, updated.GetRefined(), "refined flag must be reverted to false when criteria are insufficient")

	// An issue.unrefined outbox message must have been enqueued.
	pending, err := db.Outbox().ListPending(context.Background())
	require.NoError(t, err)

	var found bool
	for _, msg := range pending {
		if msg.Type == events.EventTypeIssueUnrefined {
			found = true
			break
		}
	}
	assert.True(t, found, "expected an issue.unrefined outbox message to be enqueued")
}

func TestHivemind_LeavesRefinedFlagUnchanged_WhenEpicEventReceived(t *testing.T) {
	s := setupRefinedEpicHandler(t)

	err := s.handler(s.ctx, events.IssueRefinedPayload{
		IssueID: s.issue.GetId(),
		ActorID: uuid.New(),
	})

	// No error must be produced.
	assert.NoError(t, err)

	// The refined flag must remain true — the handler must not touch it.
	updated, fetchErr := s.db.Issues().GetByID(context.Background(), s.issue.GetId())
	require.NoError(t, fetchErr)
	require.NotNil(t, updated)
	assert.True(t, updated.GetRefined(), "refined flag must remain true for epic-type issues")

	// No issue.unrefined outbox message must be enqueued.
	pending, outboxErr := s.db.Outbox().ListPending(context.Background())
	require.NoError(t, outboxErr)
	for _, msg := range pending {
		assert.NotEqual(t, events.EventTypeIssueUnrefined, msg.Type, "no issue.unrefined event must be enqueued for epic-type issues")
	}
}

func TestHivemind_PostsGapComment_WhenCriteriaAreVague(t *testing.T) {
	db := inmemory.NewDbContext()
	ctx := testutil.ContextWithDb(db)

	projectID := uuid.New()
	issue := newRefinedTaskWithDescription(projectID, "It should work better.")
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	gen := &vagueScenarioGenerator{}
	handler := events.HandleIssueRefinedForHivemind(gen)

	err := handler(ctx, events.IssueRefinedPayload{
		IssueID: issue.GetId(),
		ActorID: uuid.New(),
	})
	require.NoError(t, err)

	comments, total, err := db.Comments().List(context.Background(), issue.GetId(), 0, 0)
	require.NoError(t, err)
	require.Equal(t, 1, total, "expected exactly one gap comment to be posted by Hivemind")
	require.Len(t, comments, 1)

	comment := comments[0]

	// The gap comment must NOT contain scenario keywords.
	assert.NotContains(t, comment.GetBody(), "GIVEN")
	assert.NotContains(t, comment.GetBody(), "WHEN")
	assert.NotContains(t, comment.GetBody(), "THEN")

	// The gap comment must explain why scenarios could not be generated.
	assert.NotEmpty(t, comment.GetBody(), "gap comment body must not be empty")

	// The comment must be attributed to Hivemind.
	assert.Nil(t, comment.GetAuthorID(), "Hivemind comment must have no internal authorID")
	require.NotNil(t, comment.GetAuthorEmail())
	assert.Equal(t, "hivemind@hivetrack.internal", *comment.GetAuthorEmail())
	require.NotNil(t, comment.GetAuthorName())
	assert.True(t, strings.EqualFold("hivemind", *comment.GetAuthorName()), "author name should identify Hivemind")
}
