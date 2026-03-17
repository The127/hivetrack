package models

import (
	"time"

	"github.com/google/uuid"
)

type SprintStatus string

const (
	SprintStatusPlanning  SprintStatus = "planning"
	SprintStatusActive    SprintStatus = "active"
	SprintStatusCompleted SprintStatus = "completed"
)

type Sprint struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Name      string
	Goal      *string
	StartDate time.Time
	EndDate   time.Time
	Status    SprintStatus
	CreatedAt time.Time
}
