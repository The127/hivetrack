package events_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/the127/hivetrack/internal/events"
	"github.com/the127/hivetrack/internal/models"
)

func newUnassignedIssue() *models.Issue {
	reporterID := uuid.New()
	return models.NewIssue(
		uuid.New(), 1, models.IssueTypeTask, "Test issue",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
}

func TestHandleAutoAssign_UnassignedTodoToInProgress_ActorAssigned(t *testing.T) {
	issue := newUnassignedIssue()
	actorID := uuid.New()

	err := events.HandleAutoAssignOnStatusChange(context.Background(), events.IssueStatusChangedEvent{
		Issue:     issue,
		OldStatus: models.IssueStatusTodo,
		NewStatus: models.IssueStatusInProgress,
		ActorID:   actorID,
	})

	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{actorID}, issue.GetAssignees())
}

func TestHandleAutoAssign_AlreadyAssigned_NoChange(t *testing.T) {
	issue := newUnassignedIssue()
	existingAssignee := uuid.New()
	issue.SetAssignees([]uuid.UUID{existingAssignee})
	actorID := uuid.New()

	err := events.HandleAutoAssignOnStatusChange(context.Background(), events.IssueStatusChangedEvent{
		Issue:     issue,
		OldStatus: models.IssueStatusTodo,
		NewStatus: models.IssueStatusInProgress,
		ActorID:   actorID,
	})

	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{existingAssignee}, issue.GetAssignees())
}

func TestHandleAutoAssign_WrongTransition_NoAssign(t *testing.T) {
	issue := newUnassignedIssue()
	actorID := uuid.New()

	err := events.HandleAutoAssignOnStatusChange(context.Background(), events.IssueStatusChangedEvent{
		Issue:     issue,
		OldStatus: models.IssueStatusTodo,
		NewStatus: models.IssueStatusDone,
		ActorID:   actorID,
	})

	require.NoError(t, err)
	assert.Empty(t, issue.GetAssignees())
}
