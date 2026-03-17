package inmemory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
)

type OutboxRepository struct {
	messages map[uuid.UUID]*models.OutboxMessage
}

func NewOutboxRepository() *OutboxRepository {
	return &OutboxRepository{
		messages: make(map[uuid.UUID]*models.OutboxMessage),
	}
}

func (r *OutboxRepository) Enqueue(_ context.Context, msgType string, payload []byte) error {
	id := uuid.New()
	r.messages[id] = &models.OutboxMessage{
		ID:        id,
		Type:      msgType,
		Payload:   payload,
		Status:    models.OutboxStatusPending,
		CreatedAt: time.Now(),
	}
	return nil
}

func (r *OutboxRepository) ListPending(_ context.Context) ([]*models.OutboxMessage, error) {
	var result []*models.OutboxMessage
	for _, m := range r.messages {
		if m.Status == models.OutboxStatusPending {
			cp := *m
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (r *OutboxRepository) MarkDelivered(_ context.Context, id uuid.UUID) error {
	m, ok := r.messages[id]
	if !ok {
		return fmt.Errorf("outbox message %s not found: %w", id, models.ErrNotFound)
	}
	now := time.Now()
	m.Status = models.OutboxStatusDelivered
	m.DeliveredAt = &now
	return nil
}

func (r *OutboxRepository) MarkFailed(_ context.Context, id uuid.UUID, errMsg string) error {
	m, ok := r.messages[id]
	if !ok {
		return fmt.Errorf("outbox message %s not found: %w", id, models.ErrNotFound)
	}
	m.Status = models.OutboxStatusFailed
	m.Error = &errMsg
	return nil
}
