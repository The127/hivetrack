package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/change"
)

type MilestoneChange int

const (
	MilestoneChangeTitle       MilestoneChange = iota
	MilestoneChangeDescription MilestoneChange = iota
	MilestoneChangeTargetDate  MilestoneChange = iota
	MilestoneChangeClosedAt    MilestoneChange = iota
)

type Milestone struct {
	BaseModel
	change.List[MilestoneChange]

	projectID   uuid.UUID
	title       string
	description *string
	targetDate  *time.Time
	closedAt    *time.Time
}

func NewMilestone(projectID uuid.UUID, title string, description *string, targetDate *time.Time) *Milestone {
	return &Milestone{
		BaseModel:   NewBaseModel(),
		List:        change.NewList[MilestoneChange](),
		projectID:   projectID,
		title:       title,
		description: description,
		targetDate:  targetDate,
	}
}

func NewMilestoneFromDB(id uuid.UUID, createdAt time.Time, version any,
	projectID uuid.UUID, title string, description *string,
	targetDate, closedAt *time.Time) *Milestone {
	return &Milestone{
		BaseModel:   NewBaseModelFromDB(id, createdAt, createdAt, version),
		List:        change.NewList[MilestoneChange](),
		projectID:   projectID,
		title:       title,
		description: description,
		targetDate:  targetDate,
		closedAt:    closedAt,
	}
}

func (m *Milestone) GetProjectID() uuid.UUID   { return m.projectID }
func (m *Milestone) GetTitle() string          { return m.title }
func (m *Milestone) GetDescription() *string   { return m.description }
func (m *Milestone) GetTargetDate() *time.Time { return m.targetDate }
func (m *Milestone) GetClosedAt() *time.Time   { return m.closedAt }

func (m *Milestone) SetTitle(v string) {
	if m.title == v {
		return
	}
	m.title = v
	m.TrackChange(MilestoneChangeTitle)
}

func (m *Milestone) SetDescription(v *string) {
	m.description = v
	m.TrackChange(MilestoneChangeDescription)
}

func (m *Milestone) SetTargetDate(v *time.Time) {
	m.targetDate = v
	m.TrackChange(MilestoneChangeTargetDate)
}

func (m *Milestone) SetClosedAt(v *time.Time) {
	m.closedAt = v
	m.TrackChange(MilestoneChangeClosedAt)
}
