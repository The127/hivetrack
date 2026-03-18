package events

import (
	"context"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
)

// HandleAutoAssignOnStatusChange auto-assigns the actor when an unassigned issue
// moves to in_progress. It mutates the Issue pointer directly so the assignment
// is committed in the same transaction as the status change.
func HandleAutoAssignOnStatusChange(_ context.Context, evt IssueStatusChangedEvent) error {
	if evt.NewStatus != models.IssueStatusInProgress {
		return nil
	}
	if len(evt.Issue.GetAssignees()) > 0 {
		return nil
	}
	evt.Issue.SetAssignees([]uuid.UUID{evt.ActorID})
	return nil
}
