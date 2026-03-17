package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
)

type LabelRepository struct {
	ctx *DbContext
}

func NewLabelRepository(ctx *DbContext) *LabelRepository {
	return &LabelRepository{ctx: ctx}
}

func (r *LabelRepository) Insert(ctx context.Context, l *models.Label) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO labels (id, project_id, name, color) VALUES ($1,$2,$3,$4)`,
		l.ID, l.ProjectID, l.Name, l.Color,
	)
	return err
}

func (r *LabelRepository) Update(ctx context.Context, l *models.Label) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE labels SET name=$1, color=$2 WHERE id=$3`,
		l.Name, l.Color, l.ID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("label %s: %w", l.ID, models.ErrNotFound)
	}
	return nil
}

func (r *LabelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM labels WHERE id=$1`, id)
	return err
}

func (r *LabelRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Label, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, project_id, name, color FROM labels WHERE id=$1`, id)
	var l models.Label
	err := row.Scan(&l.ID, &l.ProjectID, &l.Name, &l.Color)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning label: %w", err)
	}
	return &l, nil
}

func (r *LabelRepository) List(ctx context.Context, projectID uuid.UUID) ([]*models.Label, error) {
	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx,
		`SELECT id, project_id, name, color FROM labels WHERE project_id=$1 ORDER BY name`, projectID)
	if err != nil {
		return nil, fmt.Errorf("listing labels: %w", err)
	}
	defer rows.Close()

	var labels []*models.Label
	for rows.Next() {
		var l models.Label
		if err := rows.Scan(&l.ID, &l.ProjectID, &l.Name, &l.Color); err != nil {
			return nil, fmt.Errorf("scanning label: %w", err)
		}
		labels = append(labels, &l)
	}
	return labels, rows.Err()
}
