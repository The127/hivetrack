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

type SprintRepository struct {
	ctx *DbContext
}

func NewSprintRepository(ctx *DbContext) *SprintRepository {
	return &SprintRepository{ctx: ctx}
}

func (r *SprintRepository) Insert(sprint *models.Sprint) {
	r.ctx.changeTracker.Add(change.NewEntry(sprintEntityType, sprint, change.Added))
}

func (r *SprintRepository) Update(sprint *models.Sprint) {
	r.ctx.changeTracker.Add(change.NewEntry(sprintEntityType, sprint, change.Updated))
}

func (r *SprintRepository) Delete(sprint *models.Sprint) {
	r.ctx.changeTracker.Add(change.NewEntry(sprintEntityType, sprint, change.Deleted))
}

func (r *SprintRepository) ExecuteInsert(ctx context.Context, tx *sql.Tx, sprint *models.Sprint) error {
	var version int
	err := tx.QueryRowContext(ctx,
		`INSERT INTO sprints (id, project_id, name, goal, start_date, end_date, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING version`,
		sprint.GetId(), sprint.GetProjectID(), sprint.GetName(), sprint.GetGoal(),
		sprint.GetStartDate(), sprint.GetEndDate(), sprint.GetStatus(), sprint.GetCreatedAt(),
	).Scan(&version)
	if err != nil {
		return fmt.Errorf("inserting sprint: %w", err)
	}
	sprint.SetVersion(version)
	sprint.ClearChanges()
	return nil
}

func (r *SprintRepository) ExecuteUpdate(ctx context.Context, tx *sql.Tx, sprint *models.Sprint) error {
	if !sprint.HasChanges() {
		return nil
	}

	var setClauses []string
	var args []any
	argIdx := 1

	if sprint.HasChange(models.SprintChangeName) {
		setClauses = append(setClauses, fmt.Sprintf("name=$%d", argIdx))
		args = append(args, sprint.GetName())
		argIdx++
	}
	if sprint.HasChange(models.SprintChangeGoal) {
		setClauses = append(setClauses, fmt.Sprintf("goal=$%d", argIdx))
		args = append(args, sprint.GetGoal())
		argIdx++
	}
	if sprint.HasChange(models.SprintChangeStartDate) {
		setClauses = append(setClauses, fmt.Sprintf("start_date=$%d", argIdx))
		args = append(args, sprint.GetStartDate())
		argIdx++
	}
	if sprint.HasChange(models.SprintChangeEndDate) {
		setClauses = append(setClauses, fmt.Sprintf("end_date=$%d", argIdx))
		args = append(args, sprint.GetEndDate())
		argIdx++
	}
	if sprint.HasChange(models.SprintChangeStatus) {
		setClauses = append(setClauses, fmt.Sprintf("status=$%d", argIdx))
		args = append(args, sprint.GetStatus())
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "version = version + 1")

	query := fmt.Sprintf("UPDATE sprints SET %s WHERE id=$%d", strings.Join(setClauses, ", "), argIdx) //nolint:gosec
	args = append(args, sprint.GetId())
	argIdx++

	if sprint.GetVersion() != nil {
		query += fmt.Sprintf(" AND version=$%d", argIdx)
		args = append(args, sprint.GetVersion().(int))
	}
	query += " RETURNING version"

	var version int
	err := tx.QueryRowContext(ctx, query, args...).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		if sprint.GetVersion() != nil {
			return fmt.Errorf("sprint %s: %w", sprint.GetId(), models.ErrConcurrentUpdate)
		}
		return fmt.Errorf("sprint %s: %w", sprint.GetId(), models.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("updating sprint: %w", err)
	}

	sprint.SetVersion(version)
	sprint.ClearChanges()
	return nil
}

func (r *SprintRepository) ExecuteDelete(ctx context.Context, tx *sql.Tx, sprint *models.Sprint) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM sprints WHERE id=$1`, sprint.GetId())
	if err != nil {
		return fmt.Errorf("deleting sprint: %w", err)
	}
	return nil
}

func (r *SprintRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Sprint, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, project_id, name, goal, start_date, end_date, status, created_at, version FROM sprints WHERE id=$1`, id)
	return scanSprint(row)
}

func (r *SprintRepository) List(ctx context.Context, projectID uuid.UUID) ([]*models.Sprint, error) {
	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx,
		`SELECT id, project_id, name, goal, start_date, end_date, status, created_at, version FROM sprints WHERE project_id=$1 ORDER BY start_date`,
		projectID)
	if err != nil {
		return nil, fmt.Errorf("listing sprints: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var sprints []*models.Sprint
	for rows.Next() {
		var id, projectID uuid.UUID
		var name string
		var goal sql.NullString
		var startDate, endDate, createdAt time.Time
		var status models.SprintStatus
		var version int
		if err := rows.Scan(&id, &projectID, &name, &goal, &startDate, &endDate, &status, &createdAt, &version); err != nil {
			return nil, fmt.Errorf("scanning sprint: %w", err)
		}
		var goalPtr *string
		if goal.Valid {
			goalPtr = &goal.String
		}
		sprints = append(sprints, models.NewSprintFromDB(id, createdAt, version, projectID, name, goalPtr, startDate, endDate, status))
	}
	return sprints, rows.Err()
}

func scanSprint(row *sql.Row) (*models.Sprint, error) {
	var id, projectID uuid.UUID
	var name string
	var goal sql.NullString
	var startDate, endDate, createdAt time.Time
	var status models.SprintStatus
	var version int

	err := row.Scan(&id, &projectID, &name, &goal, &startDate, &endDate, &status, &createdAt, &version)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning sprint: %w", err)
	}

	var goalPtr *string
	if goal.Valid {
		goalPtr = &goal.String
	}

	return models.NewSprintFromDB(id, createdAt, version, projectID, name, goalPtr, startDate, endDate, status), nil
}
