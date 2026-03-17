package models

import (
	"time"

	"github.com/google/uuid"
)

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "pending"
	OutboxStatusDelivered OutboxStatus = "delivered"
	OutboxStatusFailed    OutboxStatus = "failed"
)

type OutboxMessage struct {
	ID          uuid.UUID
	Type        string
	Payload     []byte
	Status      OutboxStatus
	CreatedAt   time.Time
	DeliveredAt *time.Time
	Error       *string
}
