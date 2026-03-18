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
	"github.com/the127/hivetrack/internal/repositories"
)

type ProjectRepository struct {
	ctx *DbContext
}

func NewProjectRepository(ctx *DbContext) *ProjectRepository {
	return &ProjectRepository{ctx: ctx}
}

func (r *ProjectRepository) Insert(project *models.Project) {
	r.ctx.changeTracker.Add(change.NewEntry(projectEntityType, project, change.Added))
}

func (r *ProjectRepository) Update(project *models.Project) {
	r.ctx.changeTracker.Add(change.NewEntry(projectEntityType, project, change.Updated))
}

func (r *ProjectRepository) Delete(project *models.Project) {
	r.ctx.changeTracker.Add(change.NewEntry(projectEntityType, project, change.Deleted))
}

func (r *ProjectRepository) ExecuteInsert(ctx context.Context, tx *sql.Tx, project *models.Project) error {
	var version int
	err := tx.QueryRowContext(ctx,
		`INSERT INTO projects (id, slug, name, description, archetype, archived, created_by, created_at, auto_archive_done_after_days)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING version`,
		project.GetId(), project.GetSlug(), project.GetName(), project.GetDescription(),
		project.GetArchetype(), project.GetArchived(), project.GetCreatedBy(), project.GetCreatedAt(),
		project.GetAutoArchiveDoneAfterDays(),
	).Scan(&version)
	if err != nil {
		return fmt.Errorf("inserting project: %w", err)
	}
	project.SetVersion(version)

	_, err = tx.ExecContext(ctx,
		`INSERT INTO project_issue_counters (project_id, next_number) VALUES ($1, 1)`,
		project.GetId(),
	)
	if err != nil {
		return fmt.Errorf("initializing issue counter: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO project_sprint_counters (project_id, next_number) VALUES ($1, 1)`,
		project.GetId(),
	)
	if err != nil {
		return fmt.Errorf("initializing sprint counter: %w", err)
	}

	project.ClearChanges()
	return nil
}

func (r *ProjectRepository) ExecuteUpdate(ctx context.Context, tx *sql.Tx, project *models.Project) error {
	if !project.HasChanges() {
		return nil
	}

	var setClauses []string
	var args []any
	argIdx := 1

	if project.HasChange(models.ProjectChangeSlug) {
		setClauses = append(setClauses, fmt.Sprintf("slug=$%d", argIdx))
		args = append(args, project.GetSlug())
		argIdx++
	}
	if project.HasChange(models.ProjectChangeName) {
		setClauses = append(setClauses, fmt.Sprintf("name=$%d", argIdx))
		args = append(args, project.GetName())
		argIdx++
	}
	if project.HasChange(models.ProjectChangeDescription) {
		setClauses = append(setClauses, fmt.Sprintf("description=$%d", argIdx))
		args = append(args, project.GetDescription())
		argIdx++
	}
	if project.HasChange(models.ProjectChangeArchived) {
		setClauses = append(setClauses, fmt.Sprintf("archived=$%d", argIdx))
		args = append(args, project.GetArchived())
		argIdx++
	}
	if project.HasChange(models.ProjectChangeAutoArchiveDoneAfterDays) {
		setClauses = append(setClauses, fmt.Sprintf("auto_archive_done_after_days=$%d", argIdx))
		args = append(args, project.GetAutoArchiveDoneAfterDays())
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "version = version + 1")

	query := fmt.Sprintf("UPDATE projects SET %s WHERE id=$%d", strings.Join(setClauses, ", "), argIdx) //nolint:gosec
	args = append(args, project.GetId())
	argIdx++

	if project.GetVersion() != nil {
		query += fmt.Sprintf(" AND version=$%d", argIdx)
		args = append(args, project.GetVersion().(int))
	}
	query += " RETURNING version"

	var version int
	err := tx.QueryRowContext(ctx, query, args...).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		if project.GetVersion() != nil {
			return fmt.Errorf("project %s: %w", project.GetId(), models.ErrConcurrentUpdate)
		}
		return fmt.Errorf("project %s: %w", project.GetId(), models.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("updating project: %w", err)
	}

	project.SetVersion(version)
	project.ClearChanges()
	return nil
}

func (r *ProjectRepository) ExecuteDelete(ctx context.Context, tx *sql.Tx, project *models.Project) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM projects WHERE id=$1`, project.GetId())
	if err != nil {
		return fmt.Errorf("deleting project: %w", err)
	}
	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, slug, name, description, archetype, archived, created_by, created_at, auto_archive_done_after_days, version
		FROM projects WHERE id=$1`, id)
	return scanProject(row)
}

func (r *ProjectRepository) GetBySlug(ctx context.Context, slug string) (*models.Project, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, slug, name, description, archetype, archived, created_by, created_at, auto_archive_done_after_days, version
		FROM projects WHERE slug=$1`, slug)
	return scanProject(row)
}

func (r *ProjectRepository) List(ctx context.Context, filter *repositories.ProjectFilter) ([]*models.Project, error) {
	var rows *sql.Rows
	var err error

	const selectCols = `SELECT id, slug, name, description, archetype, archived, created_by, created_at, auto_archive_done_after_days, version`

	if filter.IsAdmin {
		rows, err = r.ctx.queryContext(ctx).QueryContext(ctx,
			selectCols+` FROM projects ORDER BY name`)
	} else if filter.MemberUserID != nil {
		rows, err = r.ctx.queryContext(ctx).QueryContext(ctx,
			selectCols+` FROM projects p
			JOIN project_members pm ON p.id = pm.project_id
			WHERE pm.user_id = $1
			ORDER BY p.name`, *filter.MemberUserID)
	} else {
		rows, err = r.ctx.queryContext(ctx).QueryContext(ctx,
			selectCols+` FROM projects ORDER BY name`)
	}
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var projects []*models.Project
	for rows.Next() {
		p, err := scanProjectRow(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *ProjectRepository) AddMember(ctx context.Context, member *models.ProjectMember) error {
	return r.ctx.execDirect(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO project_members (project_id, user_id, role) VALUES ($1,$2,$3)
			ON CONFLICT (project_id, user_id) DO UPDATE SET role=EXCLUDED.role`,
			member.ProjectID, member.UserID, member.Role)
		return err
	})
}

func (r *ProjectRepository) UpdateMember(ctx context.Context, member *models.ProjectMember) error {
	return r.AddMember(ctx, member)
}

func (r *ProjectRepository) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	return r.ctx.execDirect(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`DELETE FROM project_members WHERE project_id=$1 AND user_id=$2`,
			projectID, userID)
		return err
	})
}

func (r *ProjectRepository) ListMembers(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMember, error) {
	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx,
		`SELECT project_id, user_id, role FROM project_members WHERE project_id=$1`, projectID)
	if err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var members []*models.ProjectMember
	for rows.Next() {
		var m models.ProjectMember
		if err := rows.Scan(&m.ProjectID, &m.UserID, &m.Role); err != nil {
			return nil, fmt.Errorf("scanning member: %w", err)
		}
		members = append(members, &m)
	}
	return members, rows.Err()
}

func (r *ProjectRepository) GetMember(ctx context.Context, projectID, userID uuid.UUID) (*models.ProjectMember, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT project_id, user_id, role FROM project_members WHERE project_id=$1 AND user_id=$2`,
		projectID, userID)
	var m models.ProjectMember
	err := row.Scan(&m.ProjectID, &m.UserID, &m.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning member: %w", err)
	}
	return &m, nil
}

func (r *ProjectRepository) NextIssueNumber(ctx context.Context, projectID uuid.UUID) (int, error) {
	var num int
	err := r.ctx.execDirect(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`UPDATE project_issue_counters SET next_number = next_number + 1 WHERE project_id=$1 RETURNING next_number - 1`,
			projectID).Scan(&num)
	})
	return num, err
}

func (r *ProjectRepository) NextSprintNumber(ctx context.Context, projectID uuid.UUID) (int, error) {
	var num int
	err := r.ctx.execDirect(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`UPDATE project_sprint_counters SET next_number = next_number + 1 WHERE project_id=$1 RETURNING next_number - 1`,
			projectID).Scan(&num)
	})
	return num, err
}

func scanProject(row *sql.Row) (*models.Project, error) {
	var id uuid.UUID
	var slug, name string
	var desc sql.NullString
	var archetype models.ProjectArchetype
	var archived bool
	var createdBy uuid.UUID
	var createdAt time.Time
	var autoArchive sql.NullInt32
	var version int

	err := row.Scan(&id, &slug, &name, &desc, &archetype, &archived, &createdBy, &createdAt, &autoArchive, &version)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning project: %w", err)
	}

	var descPtr *string
	if desc.Valid {
		descPtr = &desc.String
	}
	var autoArchivePtr *int
	if autoArchive.Valid {
		n := int(autoArchive.Int32)
		autoArchivePtr = &n
	}

	return models.NewProjectFromDB(id, createdAt, version, slug, name, descPtr, archetype, archived, createdBy, autoArchivePtr), nil
}

func scanProjectRow(rows *sql.Rows) (*models.Project, error) {
	var id uuid.UUID
	var slug, name string
	var desc sql.NullString
	var archetype models.ProjectArchetype
	var archived bool
	var createdBy uuid.UUID
	var createdAt time.Time
	var autoArchive sql.NullInt32
	var version int

	err := rows.Scan(&id, &slug, &name, &desc, &archetype, &archived, &createdBy, &createdAt, &autoArchive, &version)
	if err != nil {
		return nil, fmt.Errorf("scanning project: %w", err)
	}

	var descPtr *string
	if desc.Valid {
		descPtr = &desc.String
	}
	var autoArchivePtr *int
	if autoArchive.Valid {
		n := int(autoArchive.Int32)
		autoArchivePtr = &n
	}

	return models.NewProjectFromDB(id, createdAt, version, slug, name, descPtr, archetype, archived, createdBy, autoArchivePtr), nil
}
