package models

import (
	"time"

	"github.com/google/uuid"
)

type Milestone struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	Title       string
	Description *string
	TargetDate  *time.Time
	ClosedAt    *time.Time
	CreatedAt   time.Time
}
