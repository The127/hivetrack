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
	Comments() CommentRepository
	Outbox() OutboxRepository

	// SaveChanges executes all queued Insert/Update/Delete in a single transaction.
	SaveChanges(ctx context.Context) error
}

// UserRepository handles user persistence.
// Upsert is direct-execute (OIDC sync pattern).
type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetBySub(ctx context.Context, sub string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Upsert(ctx context.Context, user *models.User) error
	List(ctx context.Context) ([]*models.User, error)
}

// ProjectRepository handles project persistence.
// Insert/Update/Delete queue changes; reads and member ops are direct-execute.
type ProjectRepository interface {
	Insert(project *models.Project)
	Update(project *models.Project)
	Delete(project *models.Project)

	GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error)
	GetBySlug(ctx context.Context, slug string) (*models.Project, error)
	List(ctx context.Context, filter *ProjectFilter) ([]*models.Project, error)

	// Members — direct-execute
	AddMember(ctx context.Context, member *models.ProjectMember) error
	UpdateMember(ctx context.Context, member *models.ProjectMember) error
	RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error
	GetMember(ctx context.Context, projectID, userID uuid.UUID) (*models.ProjectMember, error)
	ListMembers(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMember, error)

	// Issue counter — direct-execute
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
// Insert/Update/Delete queue changes; reads are direct-execute.
type IssueRepository interface {
	Insert(issue *models.Issue)
	Update(issue *models.Issue)
	Delete(issue *models.Issue)

	GetByID(ctx context.Context, id uuid.UUID) (*models.Issue, error)
	GetByNumber(ctx context.Context, projectID uuid.UUID, number int) (*models.Issue, error)
	List(ctx context.Context, filter *IssueFilter) ([]*models.Issue, int, error)
}

// IssueFilter supports composable issue queries.
type IssueFilter struct {
	ProjectID   *uuid.UUID
	Status      *models.IssueStatus
	Priority    *models.IssuePriority
	SprintID    *uuid.UUID
	InBacklog   *bool // true = sprint_id IS NULL
	AssigneeID  *uuid.UUID
	Triaged     *bool
	Refined     *bool
	Text        *string
	Type        *models.IssueType
	ParentID    *uuid.UUID
	HasNoParent *bool
	Limit       int
	Offset      int
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

func (f *IssueFilter) WithInBacklog(v bool) *IssueFilter {
	f.InBacklog = &v
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

func (f *IssueFilter) WithRefined(v bool) *IssueFilter {
	f.Refined = &v
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

func (f *IssueFilter) ByType(t models.IssueType) *IssueFilter {
	f.Type = &t
	return f
}

func (f *IssueFilter) ByParentID(id uuid.UUID) *IssueFilter {
	f.ParentID = &id
	return f
}

func (f *IssueFilter) WithNoParent() *IssueFilter {
	v := true
	f.HasNoParent = &v
	return f
}

// SprintRepository handles sprint persistence.
type SprintRepository interface {
	Insert(sprint *models.Sprint)
	Update(sprint *models.Sprint)
	Delete(sprint *models.Sprint)

	GetByID(ctx context.Context, id uuid.UUID) (*models.Sprint, error)
	List(ctx context.Context, projectID uuid.UUID) ([]*models.Sprint, error)
}

// MilestoneRepository handles milestone persistence.
type MilestoneRepository interface {
	Insert(milestone *models.Milestone)
	Update(milestone *models.Milestone)
	Delete(milestone *models.Milestone)

	GetByID(ctx context.Context, id uuid.UUID) (*models.Milestone, error)
	List(ctx context.Context, projectID uuid.UUID) ([]*models.Milestone, error)
}

// LabelRepository handles label persistence.
type LabelRepository interface {
	Insert(label *models.Label)
	Update(label *models.Label)
	Delete(label *models.Label)

	GetByID(ctx context.Context, id uuid.UUID) (*models.Label, error)
	List(ctx context.Context, projectID uuid.UUID) ([]*models.Label, error)
}

// CommentRepository handles comment persistence.
// Insert/Update/Delete queue changes; reads are direct-execute.
type CommentRepository interface {
	Insert(comment *models.Comment)
	Update(comment *models.Comment)
	Delete(comment *models.Comment)

	GetByID(ctx context.Context, id uuid.UUID) (*models.Comment, error)
	List(ctx context.Context, issueID uuid.UUID, limit, offset int) ([]*models.Comment, int, error)
}

// OutboxRepository handles outbox message persistence.
// All methods are direct-execute (Enqueue runs in the current transaction).
type OutboxRepository interface {
	Enqueue(ctx context.Context, msgType string, payload []byte) error
	ListPending(ctx context.Context) ([]*models.OutboxMessage, error)
	MarkDelivered(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
}
