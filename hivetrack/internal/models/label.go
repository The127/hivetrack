package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/change"
)

type LabelChange int

const (
	LabelChangeName  LabelChange = iota
	LabelChangeColor LabelChange = iota
)

type Label struct {
	BaseModel
	change.List[LabelChange]

	projectID uuid.UUID
	name      string
	color     string
}

func NewLabel(projectID uuid.UUID, name, color string) *Label {
	return &Label{
		BaseModel: NewBaseModel(),
		List:      change.NewList[LabelChange](),
		projectID: projectID,
		name:      name,
		color:     color,
	}
}

func NewLabelFromDB(id uuid.UUID, version any, projectID uuid.UUID, name, color string) *Label {
	now := time.Time{}
	return &Label{
		BaseModel: NewBaseModelFromDB(id, now, now, version),
		List:      change.NewList[LabelChange](),
		projectID: projectID,
		name:      name,
		color:     color,
	}
}

func (l *Label) GetProjectID() uuid.UUID { return l.projectID }
func (l *Label) GetName() string         { return l.name }
func (l *Label) GetColor() string        { return l.color }

func (l *Label) SetName(v string) {
	if l.name == v {
		return
	}
	l.name = v
	l.TrackChange(LabelChangeName)
}

func (l *Label) SetColor(v string) {
	if l.color == v {
		return
	}
	l.color = v
	l.TrackChange(LabelChangeColor)
}
