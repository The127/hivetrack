package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/the127/hivetrack/internal/change"
	"github.com/the127/hivetrack/internal/models"
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

func (d *DbContext) applyChange(ctx context.Context, tx *sql.Tx, entry change.Entry) error {
	switch entry.GetItemType() {
	case projectEntityType:
		proj := entry.GetItem().(*models.Project)
		switch entry.GetChangeType() {
		case change.Added:
			return d.Projects().(*ProjectRepository).ExecuteInsert(ctx, tx, proj)
		case change.Updated:
			return d.Projects().(*ProjectRepository).ExecuteUpdate(ctx, tx, proj)
		case change.Deleted:
			return d.Projects().(*ProjectRepository).ExecuteDelete(ctx, tx, proj)
		}
	case issueEntityType:
		issue := entry.GetItem().(*models.Issue)
		switch entry.GetChangeType() {
		case change.Added:
			return d.Issues().(*IssueRepository).ExecuteInsert(ctx, tx, issue)
		case change.Updated:
			return d.Issues().(*IssueRepository).ExecuteUpdate(ctx, tx, issue)
		case change.Deleted:
			return d.Issues().(*IssueRepository).ExecuteDelete(ctx, tx, issue)
		}
	case sprintEntityType:
		sprint := entry.GetItem().(*models.Sprint)
		switch entry.GetChangeType() {
		case change.Added:
			return d.Sprints().(*SprintRepository).ExecuteInsert(ctx, tx, sprint)
		case change.Updated:
			return d.Sprints().(*SprintRepository).ExecuteUpdate(ctx, tx, sprint)
		case change.Deleted:
			return d.Sprints().(*SprintRepository).ExecuteDelete(ctx, tx, sprint)
		}
	case milestoneEntityType:
		ms := entry.GetItem().(*models.Milestone)
		switch entry.GetChangeType() {
		case change.Added:
			return d.Milestones().(*MilestoneRepository).ExecuteInsert(ctx, tx, ms)
		case change.Updated:
			return d.Milestones().(*MilestoneRepository).ExecuteUpdate(ctx, tx, ms)
		case change.Deleted:
			return d.Milestones().(*MilestoneRepository).ExecuteDelete(ctx, tx, ms)
		}
	case labelEntityType:
		lbl := entry.GetItem().(*models.Label)
		switch entry.GetChangeType() {
		case change.Added:
			return d.Labels().(*LabelRepository).ExecuteInsert(ctx, tx, lbl)
		case change.Updated:
			return d.Labels().(*LabelRepository).ExecuteUpdate(ctx, tx, lbl)
		case change.Deleted:
			return d.Labels().(*LabelRepository).ExecuteDelete(ctx, tx, lbl)
		}
	case commentEntityType:
		cmt := entry.GetItem().(*models.Comment)
		switch entry.GetChangeType() {
		case change.Added:
			return d.Comments().(*CommentRepository).ExecuteInsert(ctx, tx, cmt)
		case change.Updated:
			return d.Comments().(*CommentRepository).ExecuteUpdate(ctx, tx, cmt)
		case change.Deleted:
			return d.Comments().(*CommentRepository).ExecuteDelete(ctx, tx, cmt)
		}
	}
	return nil
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
