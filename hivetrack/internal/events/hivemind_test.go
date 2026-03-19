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

const cannedScenarios = `GIVEN a user is logged in
WHEN they click submit
THEN the form is saved`

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
