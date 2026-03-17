package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/the127/hivetrack/internal/repositories"
)

// DbContext wraps a postgres transaction and implements repositories.DbContext.
type DbContext struct {
	db *sql.DB
	tx *sql.Tx

	users      *UserRepository
	projects   *ProjectRepository
	issues     *IssueRepository
	sprints    *SprintRepository
	milestones *MilestoneRepository
	labels     *LabelRepository
	outbox     *OutboxRepository
}

// NewDbContext creates a new DbContext with a lazy transaction.
func NewDbContext(db *sql.DB) *DbContext {
	return &DbContext{db: db}
}

func (d *DbContext) getTx(ctx context.Context) (*sql.Tx, error) {
	if d.tx != nil {
		return d.tx, nil
	}
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	d.tx = tx
	return tx, nil
}

func (d *DbContext) Users() repositories.UserRepository {
	if d.users == nil {
		d.users = NewUserRepository(d)
	}
	return d.users
}

func (d *DbContext) Projects() repositories.ProjectRepository {
	if d.projects == nil {
		d.projects = NewProjectRepository(d)
	}
	return d.projects
}

func (d *DbContext) Issues() repositories.IssueRepository {
	if d.issues == nil {
		d.issues = NewIssueRepository(d)
	}
	return d.issues
}

func (d *DbContext) Sprints() repositories.SprintRepository {
	if d.sprints == nil {
		d.sprints = NewSprintRepository(d)
	}
	return d.sprints
}

func (d *DbContext) Milestones() repositories.MilestoneRepository {
	if d.milestones == nil {
		d.milestones = NewMilestoneRepository(d)
	}
	return d.milestones
}

func (d *DbContext) Labels() repositories.LabelRepository {
	if d.labels == nil {
		d.labels = NewLabelRepository(d)
	}
	return d.labels
}

func (d *DbContext) Outbox() repositories.OutboxRepository {
	if d.outbox == nil {
		d.outbox = NewOutboxRepository(d)
	}
	return d.outbox
}

func (d *DbContext) Commit(ctx context.Context) error {
	if d.tx == nil {
		return nil
	}
	if err := d.tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	d.tx = nil
	return nil
}

func (d *DbContext) Rollback(_ context.Context) error {
	if d.tx == nil {
		return nil
	}
	if err := d.tx.Rollback(); err != nil && err != sql.ErrTxDone {
		return fmt.Errorf("rolling back transaction: %w", err)
	}
	d.tx = nil
	return nil
}

// queryContext returns the current transaction if one exists, otherwise the db pool.
// Used for read operations that don't need to be in the write transaction.
func (d *DbContext) queryContext(ctx context.Context) queryRunner {
	if d.tx != nil {
		return d.tx
	}
	return d.db
}

// execContext starts a transaction if one doesn't exist, then returns it.
func (d *DbContext) execContext(ctx context.Context) (*sql.Tx, error) {
	return d.getTx(ctx)
}

// queryRunner is an abstraction over *sql.DB and *sql.Tx for read ops.
type queryRunner interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
