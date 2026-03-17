package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID
	Sub         string
	Email       string
	DisplayName string
	AvatarURL   *string
	IsAdmin     bool
	CreatedAt   time.Time
	LastLoginAt time.Time
}
