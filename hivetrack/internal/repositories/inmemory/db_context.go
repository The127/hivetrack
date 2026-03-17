package inmemory

import (
	"context"

	"github.com/the127/hivetrack/internal/repositories"
)

// DbContext is the in-memory implementation of repositories.DbContext for tests.
type DbContext struct {
	users      *UserRepository
	projects   *ProjectRepository
	issues     *IssueRepository
	sprints    *SprintRepository
	milestones *MilestoneRepository
	labels     *LabelRepository
	outbox     *OutboxRepository
}

func NewDbContext() *DbContext {
	return &DbContext{
		users:      NewUserRepository(),
		projects:   NewProjectRepository(),
		issues:     NewIssueRepository(),
		sprints:    NewSprintRepository(),
		milestones: NewMilestoneRepository(),
		labels:     NewLabelRepository(),
		outbox:     NewOutboxRepository(),
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

func (d *DbContext) Outbox() repositories.OutboxRepository {
	return d.outbox
}

func (d *DbContext) Commit(_ context.Context) error {
	return nil
}

func (d *DbContext) Rollback(_ context.Context) error {
	return nil
}
