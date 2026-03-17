package models

import (
	"time"

	"github.com/google/uuid"
)

// BaseModel provides id, createdAt, updatedAt, and xmin version to all entity models.
type BaseModel struct {
	id        uuid.UUID
	createdAt time.Time
	updatedAt time.Time
	version   any // holds postgres xmin (uint32)
}

func NewBaseModel() BaseModel {
	now := time.Now()
	return BaseModel{
		id:        uuid.New(),
		createdAt: now,
		updatedAt: now,
	}
}

func NewBaseModelFromDB(id uuid.UUID, createdAt, updatedAt time.Time, version any) BaseModel {
	return BaseModel{
		id:        id,
		createdAt: createdAt,
		updatedAt: updatedAt,
		version:   version,
	}
}

func (b BaseModel) GetId() uuid.UUID        { return b.id }
func (b BaseModel) GetCreatedAt() time.Time { return b.createdAt }
func (b BaseModel) GetUpdatedAt() time.Time { return b.updatedAt }
func (b BaseModel) GetVersion() any         { return b.version }
func (b *BaseModel) SetVersion(v any)       { b.version = v }
func (b *BaseModel) SetUpdatedAt(t time.Time) { b.updatedAt = t }
