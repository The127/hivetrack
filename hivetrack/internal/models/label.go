package models

import "github.com/google/uuid"

type Label struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Name      string
	Color     string
}
