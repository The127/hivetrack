package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/change"
	"github.com/the127/hivetrack/internal/models"
)

type MilestoneRepository struct {
	ctx *DbContext
}

func NewMilestoneRepository(ctx *DbContext) *MilestoneRepository {
	return &MilestoneRepository{ctx: ctx}
}

func (r *MilestoneRepository) Insert(milestone *models.Milestone) {
	r.ctx.changeTracker.Add(change.NewEntry(milestoneEntityType, milestone, change.Added))
}

func (r *MilestoneRepository) Update(milestone *models.Milestone) {
	r.ctx.changeTracker.Add(change.NewEntry(milestoneEntityType, milestone, change.Updated))
}

func (r *MilestoneRepository) Delete(milestone *models.Milestone) {
	r.ctx.changeTracker.Add(change.NewEntry(milestoneEntityType, milestone, change.Deleted))
}

func (r *MilestoneRepository) ExecuteInsert(ctx context.Context, tx *sql.Tx, m *models.Milestone) error {
	var xmin uint32
	err := tx.QueryRowContext(ctx,
		`INSERT INTO milestones (id, project_id, title, description, target_date, closed_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING xmin`,
		m.GetId(), m.GetProjectID(), m.GetTitle(), m.GetDescription(),
		m.GetTargetDate(), m.GetClosedAt(), m.GetCreatedAt(),
	).Scan(&xmin)
	if err != nil {
		return fmt.Errorf("inserting milestone: %w", err)
	}
	m.SetVersion(xmin)
	m.ClearChanges()
	return nil
}

func (r *MilestoneRepository) ExecuteUpdate(ctx context.Context, tx *sql.Tx, m *models.Milestone) error {
	if !m.HasChanges() {
		return nil
	}

	var setClauses []string
	var args []any
	argIdx := 1

	if m.HasChange(models.MilestoneChangeTitle) {
		setClauses = append(setClauses, fmt.Sprintf("title=$%d", argIdx))
		args = append(args, m.GetTitle())
		argIdx++
	}
	if m.HasChange(models.MilestoneChangeDescription) {
		setClauses = append(setClauses, fmt.Sprintf("description=$%d", argIdx))
		args = append(args, m.GetDescription())
		argIdx++
	}
	if m.HasChange(models.MilestoneChangeTargetDate) {
		setClauses = append(setClauses, fmt.Sprintf("target_date=$%d", argIdx))
		args = append(args, m.GetTargetDate())
		argIdx++
	}
	if m.HasChange(models.MilestoneChangeClosedAt) {
		setClauses = append(setClauses, fmt.Sprintf("closed_at=$%d", argIdx))
		args = append(args, m.GetClosedAt())
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE milestones SET %s WHERE id=$%d", strings.Join(setClauses, ", "), argIdx)
	args = append(args, m.GetId())
	argIdx++

	if m.GetVersion() != nil {
		query += fmt.Sprintf(" AND xmin=$%d::xid", argIdx)
		args = append(args, m.GetVersion().(uint32))
		argIdx++
	}
	query += " RETURNING xmin"

	var xmin uint32
	err := tx.QueryRowContext(ctx, query, args...).Scan(&xmin)
	if errors.Is(err, sql.ErrNoRows) {
		if m.GetVersion() != nil {
			return fmt.Errorf("milestone %s: %w", m.GetId(), models.ErrConcurrentUpdate)
		}
		return fmt.Errorf("milestone %s: %w", m.GetId(), models.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("updating milestone: %w", err)
	}

	m.SetVersion(xmin)
	m.ClearChanges()
	return nil
}

func (r *MilestoneRepository) ExecuteDelete(ctx context.Context, tx *sql.Tx, m *models.Milestone) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM milestones WHERE id=$1`, m.GetId())
	if err != nil {
		return fmt.Errorf("deleting milestone: %w", err)
	}
	return nil
}

func (r *MilestoneRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Milestone, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, project_id, title, description, target_date, closed_at, created_at, xmin FROM milestones WHERE id=$1`, id)
	return scanMilestone(row)
}

func (r *MilestoneRepository) List(ctx context.Context, projectID uuid.UUID) ([]*models.Milestone, error) {
	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx,
		`SELECT id, project_id, title, description, target_date, closed_at, created_at, xmin FROM milestones WHERE project_id=$1 ORDER BY created_at`,
		projectID)
	if err != nil {
		return nil, fmt.Errorf("listing milestones: %w", err)
	}
	defer rows.Close()

	var milestones []*models.Milestone
	for rows.Next() {
		var id, projectID uuid.UUID
		var title string
		var desc sql.NullString
		var targetDate, closedAt sql.NullTime
		var createdAt time.Time
		var xmin uint32
		if err := rows.Scan(&id, &projectID, &title, &desc, &targetDate, &closedAt, &createdAt, &xmin); err != nil {
			return nil, fmt.Errorf("scanning milestone: %w", err)
		}
		var descPtr *string
		if desc.Valid {
			descPtr = &desc.String
		}
		var targetDatePtr *time.Time
		if targetDate.Valid {
			targetDatePtr = &targetDate.Time
		}
		var closedAtPtr *time.Time
		if closedAt.Valid {
			closedAtPtr = &closedAt.Time
		}
		milestones = append(milestones, models.NewMilestoneFromDB(id, createdAt, xmin, projectID, title, descPtr, targetDatePtr, closedAtPtr))
	}
	return milestones, rows.Err()
}

func scanMilestone(row *sql.Row) (*models.Milestone, error) {
	var id, projectID uuid.UUID
	var title string
	var desc sql.NullString
	var targetDate, closedAt sql.NullTime
	var createdAt time.Time
	var xmin uint32

	err := row.Scan(&id, &projectID, &title, &desc, &targetDate, &closedAt, &createdAt, &xmin)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning milestone: %w", err)
	}

	var descPtr *string
	if desc.Valid {
		descPtr = &desc.String
	}
	var targetDatePtr *time.Time
	if targetDate.Valid {
		targetDatePtr = &targetDate.Time
	}
	var closedAtPtr *time.Time
	if closedAt.Valid {
		closedAtPtr = &closedAt.Time
	}

	return models.NewMilestoneFromDB(id, createdAt, xmin, projectID, title, descPtr, targetDatePtr, closedAtPtr), nil
}
