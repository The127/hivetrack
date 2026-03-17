package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
)

type OutboxRepository struct {
	ctx *DbContext
}

func NewOutboxRepository(ctx *DbContext) *OutboxRepository {
	return &OutboxRepository{ctx: ctx}
}

func (r *OutboxRepository) Enqueue(ctx context.Context, msgType string, payload []byte) error {
	return r.ctx.execDirect(ctx, func(tx *sql.Tx) error {
		id := uuid.New()
		_, err := tx.ExecContext(ctx,
			`INSERT INTO outbox_messages (id, type, payload, status, created_at) VALUES ($1,$2,$3,'pending',$4)`,
			id, msgType, payload, time.Now(),
		)
		return err
	})
}

func (r *OutboxRepository) ListPending(ctx context.Context) ([]*models.OutboxMessage, error) {
	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx,
		`SELECT id, type, payload, status, created_at, delivered_at, error FROM outbox_messages WHERE status='pending' ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("listing pending outbox: %w", err)
	}
	defer rows.Close()

	var messages []*models.OutboxMessage
	for rows.Next() {
		var m models.OutboxMessage
		var deliveredAt sql.NullTime
		var errStr sql.NullString
		if err := rows.Scan(&m.ID, &m.Type, &m.Payload, &m.Status, &m.CreatedAt, &deliveredAt, &errStr); err != nil {
			return nil, fmt.Errorf("scanning outbox message: %w", err)
		}
		if deliveredAt.Valid {
			m.DeliveredAt = &deliveredAt.Time
		}
		if errStr.Valid {
			m.Error = &errStr.String
		}
		messages = append(messages, &m)
	}
	return messages, rows.Err()
}

func (r *OutboxRepository) MarkDelivered(ctx context.Context, id uuid.UUID) error {
	return r.ctx.execDirect(ctx, func(tx *sql.Tx) error {
		now := time.Now()
		_, err := tx.ExecContext(ctx,
			`UPDATE outbox_messages SET status='delivered', delivered_at=$1 WHERE id=$2`, now, id)
		return err
	})
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	return r.ctx.execDirect(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`UPDATE outbox_messages SET status='failed', error=$1 WHERE id=$2`, errMsg, id)
		return err
	})
}
