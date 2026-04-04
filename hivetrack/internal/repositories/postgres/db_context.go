package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/the127/hivetrack/internal/change"
	"github.com/the127/hivetrack/internal/repositories"
)

// DbContext wraps a postgres connection and implements repositories.DbContext.
// Insert/Update/Delete calls are queued and flushed atomically by SaveChanges.
// Direct-execute operations (AddMember, Upsert, Enqueue, etc.) open their own mini-transactions.
type DbContext struct {
	db            *sql.DB
	changeTracker *change.Tracker

	users          *UserRepository
	projects       *ProjectRepository
	issues         *IssueRepository
	sprints        *SprintRepository
	milestones     *MilestoneRepository
	labels         *LabelRepository
	comments       *CommentRepository
	outbox         *OutboxRepository
	issueStatusLog *IssueStatusLogRepository
	auditLog       *AuditLogRepository
	refinements    *RefinementRepository
}

// NewDbContext creates a new DbContext.
func NewDbContext(db *sql.DB) *DbContext {
	return &DbContext{
		db:            db,
		changeTracker: change.NewTracker(),
	}
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

func (d *DbContext) Comments() repositories.CommentRepository {
	if d.comments == nil {
		d.comments = NewCommentRepository(d)
	}
	return d.comments
}

func (d *DbContext) Outbox() repositories.OutboxRepository {
	if d.outbox == nil {
		d.outbox = NewOutboxRepository(d)
	}
	return d.outbox
}

func (d *DbContext) IssueStatusLog() repositories.IssueStatusLogRepository {
	if d.issueStatusLog == nil {
		d.issueStatusLog = NewIssueStatusLogRepository(d)
	}
	return d.issueStatusLog
}

func (d *DbContext) AuditLog() repositories.AuditLogRepository {
	if d.auditLog == nil {
		d.auditLog = NewAuditLogRepository(d)
	}
	return d.auditLog
}

func (d *DbContext) Refinements() repositories.RefinementRepository {
	if d.refinements == nil {
		d.refinements = NewRefinementRepository(d)
	}
	return d.refinements
}

// SaveChanges executes all queued Insert/Update/Delete operations in a single transaction.
func (d *DbContext) SaveChanges(ctx context.Context) error {
	entries := d.changeTracker.GetChanges()
	if len(entries) == 0 {
		return nil
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	for _, entry := range entries {
		if err := d.applyChange(ctx, tx, entry); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	d.changeTracker.Clear()
	return nil
}

// entityApplier handles Insert/Update/Delete for a single entity type.
type entityApplier struct {
	insert func(ctx context.Context, tx *sql.Tx, item any) error
	update func(ctx context.Context, tx *sql.Tx, item any) error
	delete func(ctx context.Context, tx *sql.Tx, item any) error
}

func typedApplier[T any](
	insert func(context.Context, *sql.Tx, T) error,
	update func(context.Context, *sql.Tx, T) error,
	del func(context.Context, *sql.Tx, T) error,
) entityApplier {
	return entityApplier{
		insert: func(ctx context.Context, tx *sql.Tx, item any) error { return insert(ctx, tx, item.(T)) },
		update: func(ctx context.Context, tx *sql.Tx, item any) error { return update(ctx, tx, item.(T)) },
		delete: func(ctx context.Context, tx *sql.Tx, item any) error { return del(ctx, tx, item.(T)) },
	}
}

func (d *DbContext) entityAppliers() map[int]entityApplier {
	return map[int]entityApplier{
		projectEntityType:   typedApplier(d.Projects().(*ProjectRepository).ExecuteInsert, d.Projects().(*ProjectRepository).ExecuteUpdate, d.Projects().(*ProjectRepository).ExecuteDelete),
		issueEntityType:     typedApplier(d.Issues().(*IssueRepository).ExecuteInsert, d.Issues().(*IssueRepository).ExecuteUpdate, d.Issues().(*IssueRepository).ExecuteDelete),
		sprintEntityType:    typedApplier(d.Sprints().(*SprintRepository).ExecuteInsert, d.Sprints().(*SprintRepository).ExecuteUpdate, d.Sprints().(*SprintRepository).ExecuteDelete),
		milestoneEntityType: typedApplier(d.Milestones().(*MilestoneRepository).ExecuteInsert, d.Milestones().(*MilestoneRepository).ExecuteUpdate, d.Milestones().(*MilestoneRepository).ExecuteDelete),
		labelEntityType:     typedApplier(d.Labels().(*LabelRepository).ExecuteInsert, d.Labels().(*LabelRepository).ExecuteUpdate, d.Labels().(*LabelRepository).ExecuteDelete),
		commentEntityType:   typedApplier(d.Comments().(*CommentRepository).ExecuteInsert, d.Comments().(*CommentRepository).ExecuteUpdate, d.Comments().(*CommentRepository).ExecuteDelete),
	}
}

func (d *DbContext) applyChange(ctx context.Context, tx *sql.Tx, entry change.Entry) error {
	applier, ok := d.entityAppliers()[entry.GetItemType()]
	if !ok {
		return fmt.Errorf("unknown entity type: %d", entry.GetItemType())
	}
	switch entry.GetChangeType() {
	case change.Added:
		return applier.insert(ctx, tx, entry.GetItem())
	case change.Updated:
		return applier.update(ctx, tx, entry.GetItem())
	case change.Deleted:
		return applier.delete(ctx, tx, entry.GetItem())
	default:
		return fmt.Errorf("unknown change type: %d", entry.GetChangeType())
	}
}

// execDirect runs fn in a fresh transaction. Used by direct-execute repository methods.
func (d *DbContext) execDirect(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// queryContext returns the db pool for read operations.
func (d *DbContext) queryContext(_ context.Context) queryRunner {
	return d.db
}

// queryRunner is an abstraction over *sql.DB and *sql.Tx for read ops.
type queryRunner interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
