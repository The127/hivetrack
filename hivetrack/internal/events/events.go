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

