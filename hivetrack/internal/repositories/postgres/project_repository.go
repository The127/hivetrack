package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type ProjectRepository struct {
	ctx *DbContext
}

func NewProjectRepository(ctx *DbContext) *ProjectRepository {
	return &ProjectRepository{ctx: ctx}
}

func (r *ProjectRepository) Insert(ctx context.Context, project *models.Project) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}

	q := `INSERT INTO projects (id, slug, name, description, archetype, archived, created_by, created_at, auto_archive_done_after_days)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err = tx.ExecContext(ctx, q,
		project.ID, project.Slug, project.Name, project.Description,
		project.Archetype, project.Archived, project.CreatedBy, project.CreatedAt,
		project.AutoArchiveDoneAfterDays,
	)
	if err != nil {
		return fmt.Errorf("inserting project: %w", err)
	}

	// Initialize issue counter
	_, err = tx.ExecContext(ctx, `INSERT INTO project_issue_counters (project_id, next_number) VALUES ($1, 1)`, project.ID)
	if err != nil {
		return fmt.Errorf("initializing issue counter: %w", err)
	}

	return nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *models.Project) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}

	q := `UPDATE projects SET slug=$1, name=$2, description=$3, archived=$4, auto_archive_done_after_days=$5 WHERE id=$6`
	res, err := tx.ExecContext(ctx, q,
		project.Slug, project.Name, project.Description, project.Archived,
		project.AutoArchiveDoneAfterDays, project.ID,
	)
	if err != nil {
		return fmt.Errorf("updating project: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("project %s: %w", project.ID, models.ErrNotFound)
	}
	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM projects WHERE id=$1`, id)
	return err
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, slug, name, description, archetype, archived, created_by, created_at, auto_archive_done_after_days
		FROM projects WHERE id=$1`, id)
	return scanProject(row)
}

func (r *ProjectRepository) GetBySlug(ctx context.Context, slug string) (*models.Project, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, slug, name, description, archetype, archived, created_by, created_at, auto_archive_done_after_days
		FROM projects WHERE slug=$1`, slug)
	return scanProject(row)
}

func (r *ProjectRepository) List(ctx context.Context, filter *repositories.ProjectFilter) ([]*models.Project, error) {
	var rows *sql.Rows
	var err error

	if filter.IsAdmin {
		rows, err = r.ctx.queryContext(ctx).QueryContext(ctx,
			`SELECT id, slug, name, description, archetype, archived, created_by, created_at, auto_archive_done_after_days
			FROM projects ORDER BY name`)
	} else if filter.MemberUserID != nil {
		rows, err = r.ctx.queryContext(ctx).QueryContext(ctx,
			`SELECT p.id, p.slug, p.name, p.description, p.archetype, p.archived, p.created_by, p.created_at, p.auto_archive_done_after_days
			FROM projects p
			JOIN project_members pm ON p.id = pm.project_id
			WHERE pm.user_id = $1
			ORDER BY p.name`, *filter.MemberUserID)
	} else {
		rows, err = r.ctx.queryContext(ctx).QueryContext(ctx,
			`SELECT id, slug, name, description, archetype, archived, created_by, created_at, auto_archive_done_after_days
			FROM projects ORDER BY name`)
	}
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	defer rows.Close()

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
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO project_members (project_id, user_id, role) VALUES ($1,$2,$3)
		ON CONFLICT (project_id, user_id) DO UPDATE SET role=EXCLUDED.role`,
		member.ProjectID, member.UserID, member.Role)
	return err
}

func (r *ProjectRepository) UpdateMember(ctx context.Context, member *models.ProjectMember) error {
	return r.AddMember(ctx, member)
}

func (r *ProjectRepository) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx,
		`DELETE FROM project_members WHERE project_id=$1 AND user_id=$2`, projectID, userID)
	return err
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

func (r *ProjectRepository) ListMembers(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMember, error) {
	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx,
		`SELECT project_id, user_id, role FROM project_members WHERE project_id=$1`, projectID)
	if err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}
	defer rows.Close()

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

func (r *ProjectRepository) NextIssueNumber(ctx context.Context, projectID uuid.UUID) (int, error) {
	tx, err := r.ctx.execContext(ctx)
	if err != nil {
		return 0, err
	}

	var number int
	err = tx.QueryRowContext(ctx,
		`UPDATE project_issue_counters SET next_number = next_number + 1
		WHERE project_id = $1
		RETURNING next_number - 1`, projectID).Scan(&number)
	if err != nil {
		return 0, fmt.Errorf("getting next issue number: %w", err)
	}
	return number, nil
}

func scanProject(row *sql.Row) (*models.Project, error) {
	var p models.Project
	var desc sql.NullString
	var autoArchive sql.NullInt32
	err := row.Scan(&p.ID, &p.Slug, &p.Name, &desc, &p.Archetype, &p.Archived, &p.CreatedBy, &p.CreatedAt, &autoArchive)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning project: %w", err)
	}
	if desc.Valid {
		p.Description = &desc.String
	}
	if autoArchive.Valid {
		n := int(autoArchive.Int32)
		p.AutoArchiveDoneAfterDays = &n
	}
	return &p, nil
}

func scanProjectRow(rows *sql.Rows) (*models.Project, error) {
	var p models.Project
	var desc sql.NullString
	var autoArchive sql.NullInt32
	err := rows.Scan(&p.ID, &p.Slug, &p.Name, &desc, &p.Archetype, &p.Archived, &p.CreatedBy, &p.CreatedAt, &autoArchive)
	if err != nil {
		return nil, fmt.Errorf("scanning project: %w", err)
	}
	if desc.Valid {
		p.Description = &desc.String
	}
	if autoArchive.Valid {
		n := int(autoArchive.Int32)
		p.AutoArchiveDoneAfterDays = &n
	}
	return &p, nil
}
