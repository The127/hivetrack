package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
)

type MilestoneRepository struct {
	ctx *DbContext
}

func NewMilestoneRepository(ctx *DbContext) *MilestoneRepository {
	return &MilestoneRepository{ctx: ctx}
}

func (r *MilestoneRepository) Insert(ctx context.Context, m *models.Milestone) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO milestones (id, project_id, title, description, target_date, closed_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		m.ID, m.ProjectID, m.Title, m.Description, m.TargetDate, m.ClosedAt, m.CreatedAt,
	)
	return err
}

func (r *MilestoneRepository) Update(ctx context.Context, m *models.Milestone) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE milestones SET title=$1, description=$2, target_date=$3, closed_at=$4 WHERE id=$5`,
		m.Title, m.Description, m.TargetDate, m.ClosedAt, m.ID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("milestone %s: %w", m.ID, models.ErrNotFound)
	}
	return nil
}

func (r *MilestoneRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM milestones WHERE id=$1`, id)
	return err
}

func (r *MilestoneRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Milestone, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, project_id, title, description, target_date, closed_at, created_at FROM milestones WHERE id=$1`, id)
	return scanMilestone(row)
}

func (r *MilestoneRepository) List(ctx context.Context, projectID uuid.UUID) ([]*models.Milestone, error) {
	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx,
		`SELECT id, project_id, title, description, target_date, closed_at, created_at FROM milestones WHERE project_id=$1 ORDER BY created_at`,
		projectID)
	if err != nil {
		return nil, fmt.Errorf("listing milestones: %w", err)
	}
	defer rows.Close()

	var milestones []*models.Milestone
	for rows.Next() {
		var m models.Milestone
		var desc sql.NullString
		var targetDate, closedAt sql.NullTime
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.Title, &desc, &targetDate, &closedAt, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning milestone: %w", err)
		}
		if desc.Valid {
			m.Description = &desc.String
		}
		if targetDate.Valid {
			m.TargetDate = &targetDate.Time
		}
		if closedAt.Valid {
			m.ClosedAt = &closedAt.Time
		}
		milestones = append(milestones, &m)
	}
	return milestones, rows.Err()
}

func scanMilestone(row *sql.Row) (*models.Milestone, error) {
	var m models.Milestone
	var desc sql.NullString
	var targetDate, closedAt sql.NullTime
	err := row.Scan(&m.ID, &m.ProjectID, &m.Title, &desc, &targetDate, &closedAt, &m.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning milestone: %w", err)
	}
	if desc.Valid {
		m.Description = &desc.String
	}
	if targetDate.Valid {
		m.TargetDate = &targetDate.Time
	}
	if closedAt.Valid {
		m.ClosedAt = &closedAt.Time
	}
	return &m, nil
}
