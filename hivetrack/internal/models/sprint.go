package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/change"
)

type SprintStatus string

const (
	SprintStatusPlanning  SprintStatus = "planning"
	SprintStatusActive    SprintStatus = "active"
	SprintStatusCompleted SprintStatus = "completed"
)

type SprintChange int

const (
	SprintChangeName      SprintChange = iota
	SprintChangeGoal      SprintChange = iota
	SprintChangeStartDate SprintChange = iota
	SprintChangeEndDate   SprintChange = iota
	SprintChangeStatus    SprintChange = iota
)

type Sprint struct {
	BaseModel
	change.List[SprintChange]

	projectID uuid.UUID
	name      string
	goal      *string
	startDate time.Time
	endDate   time.Time
	status    SprintStatus
}

func NewSprint(projectID uuid.UUID, name string, goal *string, startDate, endDate time.Time, status SprintStatus) *Sprint {
	return &Sprint{
		BaseModel: NewBaseModel(),
		List:      change.NewList[SprintChange](),
		projectID: projectID,
		name:      name,
		goal:      goal,
		startDate: startDate,
		endDate:   endDate,
		status:    status,
	}
}

func NewSprintFromDB(id uuid.UUID, createdAt time.Time, version any,
	projectID uuid.UUID, name string, goal *string,
	startDate, endDate time.Time, status SprintStatus) *Sprint {
	return &Sprint{
		BaseModel: NewBaseModelFromDB(id, createdAt, createdAt, version),
		List:      change.NewList[SprintChange](),
		projectID: projectID,
		name:      name,
		goal:      goal,
		startDate: startDate,
		endDate:   endDate,
		status:    status,
	}
}

func (s *Sprint) GetProjectID() uuid.UUID { return s.projectID }
func (s *Sprint) GetName() string         { return s.name }
func (s *Sprint) GetGoal() *string        { return s.goal }
func (s *Sprint) GetStartDate() time.Time { return s.startDate }
func (s *Sprint) GetEndDate() time.Time   { return s.endDate }
func (s *Sprint) GetStatus() SprintStatus { return s.status }

func (s *Sprint) SetName(v string) {
	if s.name == v {
		return
	}
	s.name = v
	s.TrackChange(SprintChangeName)
}

func (s *Sprint) SetGoal(v *string) {
	s.goal = v
	s.TrackChange(SprintChangeGoal)
}

func (s *Sprint) SetStartDate(v time.Time) {
	s.startDate = v
	s.TrackChange(SprintChangeStartDate)
}

func (s *Sprint) SetEndDate(v time.Time) {
	s.endDate = v
	s.TrackChange(SprintChangeEndDate)
}

func (s *Sprint) SetStatus(v SprintStatus) {
	if s.status == v {
		return
	}
	s.status = v
	s.TrackChange(SprintChangeStatus)
}
