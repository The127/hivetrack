package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditLogEntry records a single auditable action on an entity.
type AuditLogEntry struct {
	ID         uuid.UUID
	EntityID   uuid.UUID
	Action     string
	ActorID    uuid.UUID
	RecordedAt time.Time
}

func NewAuditLogEntry(entityID uuid.UUID, action string, actorID uuid.UUID) *AuditLogEntry {
	return &AuditLogEntry{
		ID:         uuid.New(),
		EntityID:   entityID,
		Action:     action,
		ActorID:    actorID,
		RecordedAt: time.Now(),
	}
}
