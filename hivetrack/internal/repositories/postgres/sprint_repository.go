package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
)

type SprintRepository struct {
	ctx *DbContext
}

func NewSprintRepository(ctx *DbContext) *SprintRepository {
	return &SprintRepository{ctx: ctx}
}

func (r *SprintRepository) Insert(ctx context.Context, sprint *models.Sprint) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO sprints (id, project_id, name, goal, start_date, end_date, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		sprint.ID, sprint.ProjectID, sprint.Name, sprint.Goal,
		sprint.StartDate, sprint.EndDate, sprint.Status, sprint.CreatedAt,
	)
	return err
}

func (r *SprintRepository) Update(ctx context.Context, sprint *models.Sprint) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE sprints SET name=$1, goal=$2, start_date=$3, end_date=$4, status=$5 WHERE id=$6`,
		sprint.Name, sprint.Goal, sprint.StartDate, sprint.EndDate, sprint.Status, sprint.ID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("sprint %s: %w", sprint.ID, models.ErrNotFound)
	}
	return nil
}

func (r *SprintRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM sprints WHERE id=$1`, id)
	return err
}

func (r *SprintRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Sprint, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, project_id, name, goal, start_date, end_date, status, created_at FROM sprints WHERE id=$1`, id)
	return scanSprint(row)
}

func (r *SprintRepository) List(ctx context.Context, projectID uuid.UUID) ([]*models.Sprint, error) {
	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx,
		`SELECT id, project_id, name, goal, start_date, end_date, status, created_at FROM sprints WHERE project_id=$1 ORDER BY start_date`,
		projectID)
	if err != nil {
		return nil, fmt.Errorf("listing sprints: %w", err)
	}
	defer rows.Close()

	var sprints []*models.Sprint
	for rows.Next() {
		var s models.Sprint
		var goal sql.NullString
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Name, &goal, &s.StartDate, &s.EndDate, &s.Status, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning sprint: %w", err)
		}
		if goal.Valid {
			s.Goal = &goal.String
		}
		sprints = append(sprints, &s)
	}
	return sprints, rows.Err()
}

func scanSprint(row *sql.Row) (*models.Sprint, error) {
	var s models.Sprint
	var goal sql.NullString
	err := row.Scan(&s.ID, &s.ProjectID, &s.Name, &goal, &s.StartDate, &s.EndDate, &s.Status, &s.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning sprint: %w", err)
	}
	if goal.Valid {
		s.Goal = &goal.String
	}
	return &s, nil
}
