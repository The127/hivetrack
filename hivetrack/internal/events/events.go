// Package events contains domain event types and outbox delivery infrastructure.
package events

import (
	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
)

// IssueStatusChangedEvent is fired when an issue's status is changed.
type IssueStatusChangedEvent struct {
	Issue     *models.Issue
	OldStatus models.IssueStatus
	NewStatus models.IssueStatus
	ActorID   uuid.UUID
}

// EventTypeIssueRefined is the outbox message type for issue.refined events.
const EventTypeIssueRefined = "issue.refined"

// IssueRefinedPayload is the outbox payload for an issue.refined event.
type IssueRefinedPayload struct {
	IssueID uuid.UUID `json:"issue_id"`
	ActorID uuid.UUID `json:"actor_id"`
}

// EventTypeIssueUnrefined is the outbox message type for issue.unrefined events.
const EventTypeIssueUnrefined = "issue.unrefined"

// IssueUnrefinedPayload is the outbox payload for an issue.unrefined event.
type IssueUnrefinedPayload struct {
	IssueID   uuid.UUID `json:"issue_id"`
	ProjectID uuid.UUID `json:"project_id"`
}
