package inmemory

import (
	"context"
	"fmt"

	"github.com/the127/hivetrack/internal/change"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

// DbContext is the in-memory implementation of repositories.DbContext for tests.
type DbContext struct {
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
}

func NewDbContext() *DbContext {
	tracker := change.NewTracker()
	issueRepo := NewIssueRepository(tracker)
	return &DbContext{
		changeTracker:  tracker,
		users:          NewUserRepository(),
		projects:       NewProjectRepository(tracker),
		issues:         issueRepo,
		sprints:        NewSprintRepository(tracker),
		milestones:     NewMilestoneRepository(tracker, issueRepo),
		labels:         NewLabelRepository(tracker),
		comments:       NewCommentRepository(tracker),
		outbox:         NewOutboxRepository(),
		issueStatusLog: NewIssueStatusLogRepository(issueRepo),
	}
}

func (d *DbContext) Users() repositories.UserRepository {
	return d.users
}

func (d *DbContext) Projects() repositories.ProjectRepository {
	return d.projects
}

func (d *DbContext) Issues() repositories.IssueRepository {
	return d.issues
}

func (d *DbContext) Sprints() repositories.SprintRepository {
	return d.sprints
}

func (d *DbContext) Milestones() repositories.MilestoneRepository {
	return d.milestones
}

func (d *DbContext) Labels() repositories.LabelRepository {
	return d.labels
}

func (d *DbContext) Comments() repositories.CommentRepository {
	return d.comments
}

func (d *DbContext) Outbox() repositories.OutboxRepository {
	return d.outbox
}

func (d *DbContext) IssueStatusLog() repositories.IssueStatusLogRepository {
	return d.issueStatusLog
}

// SaveChanges applies all queued Insert/Update/Delete operations to the in-memory stores.
func (d *DbContext) SaveChanges(_ context.Context) error {
	for _, entry := range d.changeTracker.GetChanges() {
		switch item := entry.GetItem().(type) {
		case *models.Project:
			switch entry.GetChangeType() {
			case change.Added:
				if _, exists := d.projects.bySlug[item.GetSlug()]; exists {
					return fmt.Errorf("project with slug %q already exists: %w", item.GetSlug(), models.ErrConflict)
				}
				d.projects.byID[item.GetId()] = item
				d.projects.bySlug[item.GetSlug()] = item
				d.projects.counters[item.GetId()] = 1
				item.ClearChanges()
			case change.Updated:
				existing, ok := d.projects.byID[item.GetId()]
				if !ok {
					return fmt.Errorf("project %s: %w", item.GetId(), models.ErrNotFound)
				}
				if existing.GetSlug() != item.GetSlug() {
					delete(d.projects.bySlug, existing.GetSlug())
				}
				d.projects.byID[item.GetId()] = item
				d.projects.bySlug[item.GetSlug()] = item
				item.ClearChanges()
			case change.Deleted:
				existing := d.projects.byID[item.GetId()]
				if existing != nil {
					delete(d.projects.bySlug, existing.GetSlug())
				}
				delete(d.projects.byID, item.GetId())
				delete(d.projects.members, item.GetId())
				delete(d.projects.counters, item.GetId())
			}
		case *models.Issue:
			switch entry.GetChangeType() {
			case change.Added:
				key := issueKey(item.GetProjectID(), item.GetNumber())
				if _, exists := d.issues.byProjectAndNumber[key]; exists {
					return fmt.Errorf("issue %s already exists: %w", key, models.ErrConflict)
				}
				d.issues.byID[item.GetId()] = item
				d.issues.byProjectAndNumber[key] = item
				item.ClearChanges()
			case change.Updated:
				if _, ok := d.issues.byID[item.GetId()]; !ok {
					return fmt.Errorf("issue %s: %w", item.GetId(), models.ErrNotFound)
				}
				d.issues.byID[item.GetId()] = item
				d.issues.byProjectAndNumber[issueKey(item.GetProjectID(), item.GetNumber())] = item
				item.ClearChanges()
			case change.Deleted:
				existing := d.issues.byID[item.GetId()]
				if existing != nil {
					delete(d.issues.byProjectAndNumber, issueKey(existing.GetProjectID(), existing.GetNumber()))
				}
				delete(d.issues.byID, item.GetId())
			}
		case *models.Sprint:
			switch entry.GetChangeType() {
			case change.Added:
				d.sprints.byID[item.GetId()] = item
				item.ClearChanges()
			case change.Updated:
				if _, ok := d.sprints.byID[item.GetId()]; !ok {
					return fmt.Errorf("sprint %s: %w", item.GetId(), models.ErrNotFound)
				}
				d.sprints.byID[item.GetId()] = item
				item.ClearChanges()
			case change.Deleted:
				delete(d.sprints.byID, item.GetId())
			}
		case *models.Milestone:
			switch entry.GetChangeType() {
			case change.Added:
				d.milestones.byID[item.GetId()] = item
				item.ClearChanges()
			case change.Updated:
				if _, ok := d.milestones.byID[item.GetId()]; !ok {
					return fmt.Errorf("milestone %s: %w", item.GetId(), models.ErrNotFound)
				}
				d.milestones.byID[item.GetId()] = item
				item.ClearChanges()
			case change.Deleted:
				delete(d.milestones.byID, item.GetId())
			}
		case *models.Label:
			switch entry.GetChangeType() {
			case change.Added:
				d.labels.byID[item.GetId()] = item
				item.ClearChanges()
			case change.Updated:
				if _, ok := d.labels.byID[item.GetId()]; !ok {
					return fmt.Errorf("label %s: %w", item.GetId(), models.ErrNotFound)
				}
				d.labels.byID[item.GetId()] = item
				item.ClearChanges()
			case change.Deleted:
				delete(d.labels.byID, item.GetId())
			}
		case *models.Comment:
			switch entry.GetChangeType() {
			case change.Added:
				d.comments.byID[item.GetId()] = item
				item.ClearChanges()
			case change.Updated:
				if _, ok := d.comments.byID[item.GetId()]; !ok {
					return fmt.Errorf("comment %s: %w", item.GetId(), models.ErrNotFound)
				}
				d.comments.byID[item.GetId()] = item
				item.ClearChanges()
			case change.Deleted:
				delete(d.comments.byID, item.GetId())
			}
		}
	}
	d.changeTracker.Clear()
	return nil
}
