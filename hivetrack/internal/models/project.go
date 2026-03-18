package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/change"
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

type ProjectChange int

const (
	ProjectChangeSlug                     ProjectChange = iota
	ProjectChangeName                     ProjectChange = iota
	ProjectChangeDescription              ProjectChange = iota
	ProjectChangeArchived                 ProjectChange = iota
	ProjectChangeAutoArchiveDoneAfterDays ProjectChange = iota
)

type Project struct {
	BaseModel
	change.List[ProjectChange]

	slug                     string
	name                     string
	description              *string
	archetype                ProjectArchetype
	archived                 bool
	createdBy                uuid.UUID
	autoArchiveDoneAfterDays *int
}

func NewProject(createdBy uuid.UUID, slug, name string, archetype ProjectArchetype) *Project {
	return &Project{
		BaseModel: NewBaseModel(),
		List:      change.NewList[ProjectChange](),
		slug:      slug,
		name:      name,
		archetype: archetype,
		createdBy: createdBy,
	}
}

func NewProjectFromDB(id uuid.UUID, createdAt time.Time, version any,
	slug, name string, description *string, archetype ProjectArchetype,
	archived bool, createdBy uuid.UUID, autoArchiveDoneAfterDays *int) *Project {
	return &Project{
		BaseModel:                NewBaseModelFromDB(id, createdAt, createdAt, version),
		List:                     change.NewList[ProjectChange](),
		slug:                     slug,
		name:                     name,
		description:              description,
		archetype:                archetype,
		archived:                 archived,
		createdBy:                createdBy,
		autoArchiveDoneAfterDays: autoArchiveDoneAfterDays,
	}
}

func (p *Project) GetSlug() string                   { return p.slug }
func (p *Project) GetName() string                   { return p.name }
func (p *Project) GetDescription() *string           { return p.description }
func (p *Project) GetArchetype() ProjectArchetype    { return p.archetype }
func (p *Project) GetArchived() bool                 { return p.archived }
func (p *Project) GetCreatedBy() uuid.UUID           { return p.createdBy }
func (p *Project) GetAutoArchiveDoneAfterDays() *int { return p.autoArchiveDoneAfterDays }

func (p *Project) SetSlug(v string) {
	if p.slug == v {
		return
	}
	p.slug = v
	p.TrackChange(ProjectChangeSlug)
}

func (p *Project) SetName(v string) {
	if p.name == v {
		return
	}
	p.name = v
	p.TrackChange(ProjectChangeName)
}

func (p *Project) SetDescription(v *string) {
	p.description = v
	p.TrackChange(ProjectChangeDescription)
}

func (p *Project) SetArchived(v bool) {
	if p.archived == v {
		return
	}
	p.archived = v
	p.TrackChange(ProjectChangeArchived)
}

func (p *Project) SetAutoArchiveDoneAfterDays(v *int) {
	p.autoArchiveDoneAfterDays = v
	p.TrackChange(ProjectChangeAutoArchiveDoneAfterDays)
}

// ProjectMember is a join table record — not a tracked entity.
type ProjectMember struct {
	ProjectID uuid.UUID
	UserID    uuid.UUID
	Role      ProjectRole
}
