package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
)

// DbContext is the unit of work. All repository access goes through it.
type DbContext interface {
	Users() UserRepository
	Projects() ProjectRepository
	Issues() IssueRepository
	Sprints() SprintRepository
	Milestones() MilestoneRepository
	Labels() LabelRepository
	Outbox() OutboxRepository

	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// UserRepository handles user persistence.
type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetBySub(ctx context.Context, sub string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Upsert(ctx context.Context, user *models.User) error
	List(ctx context.Context) ([]*models.User, error)
}

// ProjectRepository handles project persistence.
type ProjectRepository interface {
	Insert(ctx context.Context, project *models.Project) error
	Update(ctx context.Context, project *models.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error)
	GetBySlug(ctx context.Context, slug string) (*models.Project, error)
	List(ctx context.Context, filter *ProjectFilter) ([]*models.Project, error)

	// Members
	AddMember(ctx context.Context, member *models.ProjectMember) error
	UpdateMember(ctx context.Context, member *models.ProjectMember) error
	RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error
	GetMember(ctx context.Context, projectID, userID uuid.UUID) (*models.ProjectMember, error)
	ListMembers(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMember, error)

	// Issue counter
	NextIssueNumber(ctx context.Context, projectID uuid.UUID) (int, error)
}

// ProjectFilter filters project list queries.
type ProjectFilter struct {
	MemberUserID *uuid.UUID
	IsAdmin      bool
}

func NewProjectFilter() *ProjectFilter {
	return &ProjectFilter{}
}

func (f *ProjectFilter) ForMember(userID uuid.UUID) *ProjectFilter {
	f.MemberUserID = &userID
	return f
}

func (f *ProjectFilter) AsAdmin() *ProjectFilter {
	f.IsAdmin = true
	return f
}

// IssueRepository handles issue persistence.
type IssueRepository interface {
	Insert(ctx context.Context, issue *models.Issue) error
	Update(ctx context.Context, issue *models.Issue) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Issue, error)
	GetByNumber(ctx context.Context, projectID uuid.UUID, number int) (*models.Issue, error)
	List(ctx context.Context, filter *IssueFilter) ([]*models.Issue, int, error)
}

// IssueFilter supports composable issue queries.
type IssueFilter struct {
	ProjectID  *uuid.UUID
	Status     *models.IssueStatus
	Priority   *models.IssuePriority
	SprintID   *uuid.UUID
	AssigneeID *uuid.UUID
	Triaged    *bool
	Text       *string
	Limit      int
	Offset     int
}

func NewIssueFilter() *IssueFilter {
	return &IssueFilter{}
}

func (f *IssueFilter) ByProjectID(id uuid.UUID) *IssueFilter {
	f.ProjectID = &id
	return f
}

func (f *IssueFilter) ByStatus(s models.IssueStatus) *IssueFilter {
	f.Status = &s
	return f
}

func (f *IssueFilter) ByPriority(p models.IssuePriority) *IssueFilter {
	f.Priority = &p
	return f
}

func (f *IssueFilter) BySprintID(id uuid.UUID) *IssueFilter {
	f.SprintID = &id
	return f
}

func (f *IssueFilter) ByAssigneeID(id uuid.UUID) *IssueFilter {
	f.AssigneeID = &id
	return f
}

func (f *IssueFilter) WithTriaged(v bool) *IssueFilter {
	f.Triaged = &v
	return f
}

func (f *IssueFilter) WithText(t string) *IssueFilter {
	f.Text = &t
	return f
}

func (f *IssueFilter) WithPagination(limit, offset int) *IssueFilter {
	f.Limit = limit
	f.Offset = offset
	return f
}

// SprintRepository handles sprint persistence.
type SprintRepository interface {
	Insert(ctx context.Context, sprint *models.Sprint) error
	Update(ctx context.Context, sprint *models.Sprint) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Sprint, error)
	List(ctx context.Context, projectID uuid.UUID) ([]*models.Sprint, error)
}

// MilestoneRepository handles milestone persistence.
type MilestoneRepository interface {
	Insert(ctx context.Context, milestone *models.Milestone) error
	Update(ctx context.Context, milestone *models.Milestone) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Milestone, error)
	List(ctx context.Context, projectID uuid.UUID) ([]*models.Milestone, error)
}

// LabelRepository handles label persistence.
type LabelRepository interface {
	Insert(ctx context.Context, label *models.Label) error
	Update(ctx context.Context, label *models.Label) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Label, error)
	List(ctx context.Context, projectID uuid.UUID) ([]*models.Label, error)
}

// OutboxRepository handles outbox message persistence.
type OutboxRepository interface {
	Enqueue(ctx context.Context, msgType string, payload []byte) error
	ListPending(ctx context.Context) ([]*models.OutboxMessage, error)
	MarkDelivered(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
}
