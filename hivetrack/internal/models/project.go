package models

import (
	"time"

	"github.com/google/uuid"
)

type ProjectArchetype string

const (
	ProjectArchetypeSoftware ProjectArchetype = "software"
	ProjectArchetypeSupport  ProjectArchetype = "support"
)

type ProjectRole string

const (
	ProjectRoleAdmin  ProjectRole = "project_admin"
	ProjectRoleMember ProjectRole = "project_member"
	ProjectRoleViewer ProjectRole = "viewer"
)

type Project struct {
	ID          uuid.UUID
	Slug        string
	Name        string
	Description *string
	Archetype   ProjectArchetype
	Archived    bool
	CreatedBy   uuid.UUID
	CreatedAt   time.Time

	AutoArchiveDoneAfterDays *int
}

type ProjectMember struct {
	ProjectID uuid.UUID
	UserID    uuid.UUID
	Role      ProjectRole
}
