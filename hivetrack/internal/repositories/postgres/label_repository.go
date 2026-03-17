package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/change"
	"github.com/the127/hivetrack/internal/models"
)

type LabelRepository struct {
	ctx *DbContext
}

func NewLabelRepository(ctx *DbContext) *LabelRepository {
	return &LabelRepository{ctx: ctx}
}

func (r *LabelRepository) Insert(label *models.Label) {
	r.ctx.changeTracker.Add(change.NewEntry(labelEntityType, label, change.Added))
}

func (r *LabelRepository) Update(label *models.Label) {
	r.ctx.changeTracker.Add(change.NewEntry(labelEntityType, label, change.Updated))
}

func (r *LabelRepository) Delete(label *models.Label) {
	r.ctx.changeTracker.Add(change.NewEntry(labelEntityType, label, change.Deleted))
}

func (r *LabelRepository) ExecuteInsert(ctx context.Context, tx *sql.Tx, l *models.Label) error {
	var xmin uint32
	err := tx.QueryRowContext(ctx,
		`INSERT INTO labels (id, project_id, name, color) VALUES ($1,$2,$3,$4) RETURNING xmin`,
		l.GetId(), l.GetProjectID(), l.GetName(), l.GetColor(),
	).Scan(&xmin)
	if err != nil {
		return fmt.Errorf("inserting label: %w", err)
	}
	l.SetVersion(xmin)
	l.ClearChanges()
	return nil
}

func (r *LabelRepository) ExecuteUpdate(ctx context.Context, tx *sql.Tx, l *models.Label) error {
	if !l.HasChanges() {
		return nil
	}

	var setClauses []string
	var args []any
	argIdx := 1

	if l.HasChange(models.LabelChangeName) {
		setClauses = append(setClauses, fmt.Sprintf("name=$%d", argIdx))
		args = append(args, l.GetName())
		argIdx++
	}
	if l.HasChange(models.LabelChangeColor) {
		setClauses = append(setClauses, fmt.Sprintf("color=$%d", argIdx))
		args = append(args, l.GetColor())
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE labels SET %s WHERE id=$%d", strings.Join(setClauses, ", "), argIdx)
	args = append(args, l.GetId())
	argIdx++

	if l.GetVersion() != nil {
		query += fmt.Sprintf(" AND xmin=$%d::xid", argIdx)
		args = append(args, l.GetVersion().(uint32))
		argIdx++
	}
	query += " RETURNING xmin"

	var xmin uint32
	err := tx.QueryRowContext(ctx, query, args...).Scan(&xmin)
	if errors.Is(err, sql.ErrNoRows) {
		if l.GetVersion() != nil {
			return fmt.Errorf("label %s: %w", l.GetId(), models.ErrConcurrentUpdate)
		}
		return fmt.Errorf("label %s: %w", l.GetId(), models.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("updating label: %w", err)
	}

	l.SetVersion(xmin)
	l.ClearChanges()
	return nil
}

func (r *LabelRepository) ExecuteDelete(ctx context.Context, tx *sql.Tx, l *models.Label) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM labels WHERE id=$1`, l.GetId())
	if err != nil {
		return fmt.Errorf("deleting label: %w", err)
	}
	return nil
}

func (r *LabelRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Label, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, project_id, name, color, xmin FROM labels WHERE id=$1`, id)
	var labelID, projectID uuid.UUID
	var name, color string
	var xmin uint32
	err := row.Scan(&labelID, &projectID, &name, &color, &xmin)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning label: %w", err)
	}
	return models.NewLabelFromDB(labelID, xmin, projectID, name, color), nil
}

func (r *LabelRepository) List(ctx context.Context, projectID uuid.UUID) ([]*models.Label, error) {
	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx,
		`SELECT id, project_id, name, color, xmin FROM labels WHERE project_id=$1 ORDER BY name`, projectID)
	if err != nil {
		return nil, fmt.Errorf("listing labels: %w", err)
	}
	defer rows.Close()

	var labels []*models.Label
	for rows.Next() {
		var labelID, lProjectID uuid.UUID
		var name, color string
		var xmin uint32
		if err := rows.Scan(&labelID, &lProjectID, &name, &color, &xmin); err != nil {
			return nil, fmt.Errorf("scanning label: %w", err)
		}
		labels = append(labels, models.NewLabelFromDB(labelID, xmin, lProjectID, name, color))
	}
	return labels, rows.Err()
}
