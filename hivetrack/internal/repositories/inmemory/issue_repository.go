package inmemory

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type IssueRepository struct {
	byID        map[uuid.UUID]*models.Issue
	byProjectAndNumber map[string]*models.Issue // "projectID:number" -> issue
}

func NewIssueRepository() *IssueRepository {
	return &IssueRepository{
		byID:               make(map[uuid.UUID]*models.Issue),
		byProjectAndNumber: make(map[string]*models.Issue),
	}
}

func issueKey(projectID uuid.UUID, number int) string {
	return fmt.Sprintf("%s:%d", projectID, number)
}

func (r *IssueRepository) Insert(_ context.Context, issue *models.Issue) error {
	key := issueKey(issue.ProjectID, issue.Number)
	if _, exists := r.byProjectAndNumber[key]; exists {
		return fmt.Errorf("issue %s already exists: %w", key, models.ErrConflict)
	}
	cp := copyIssue(issue)
	r.byID[issue.ID] = cp
	r.byProjectAndNumber[key] = cp
	return nil
}

func (r *IssueRepository) Update(_ context.Context, issue *models.Issue) error {
	if _, ok := r.byID[issue.ID]; !ok {
		return fmt.Errorf("issue %s not found: %w", issue.ID, models.ErrNotFound)
	}
	cp := copyIssue(issue)
	r.byID[issue.ID] = cp
	r.byProjectAndNumber[issueKey(issue.ProjectID, issue.Number)] = cp
	return nil
}

func (r *IssueRepository) Delete(_ context.Context, id uuid.UUID) error {
	issue, ok := r.byID[id]
	if !ok {
		return fmt.Errorf("issue %s not found: %w", id, models.ErrNotFound)
	}
	delete(r.byProjectAndNumber, issueKey(issue.ProjectID, issue.Number))
	delete(r.byID, id)
	return nil
}

func (r *IssueRepository) GetByID(_ context.Context, id uuid.UUID) (*models.Issue, error) {
	issue, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	return copyIssue(issue), nil
}

func (r *IssueRepository) GetByNumber(_ context.Context, projectID uuid.UUID, number int) (*models.Issue, error) {
	issue, ok := r.byProjectAndNumber[issueKey(projectID, number)]
	if !ok {
		return nil, nil
	}
	return copyIssue(issue), nil
}

func (r *IssueRepository) List(_ context.Context, filter *repositories.IssueFilter) ([]*models.Issue, int, error) {
	var result []*models.Issue
	for _, issue := range r.byID {
		if !matchesIssueFilter(issue, filter) {
			continue
		}
		result = append(result, copyIssue(issue))
	}
	total := len(result)

	// Apply pagination
	if filter.Offset > 0 {
		if filter.Offset >= len(result) {
			return nil, total, nil
		}
		result = result[filter.Offset:]
	}
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}

	return result, total, nil
}

func matchesIssueFilter(issue *models.Issue, filter *repositories.IssueFilter) bool {
	if filter.ProjectID != nil && issue.ProjectID != *filter.ProjectID {
		return false
	}
	if filter.Status != nil && issue.Status != *filter.Status {
		return false
	}
	if filter.Priority != nil && issue.Priority != *filter.Priority {
		return false
	}
	if filter.SprintID != nil {
		if issue.SprintID == nil || *issue.SprintID != *filter.SprintID {
			return false
		}
	}
	if filter.AssigneeID != nil {
		found := false
		for _, a := range issue.Assignees {
			if a == *filter.AssigneeID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if filter.Triaged != nil && issue.Triaged != *filter.Triaged {
		return false
	}
	if filter.Text != nil && *filter.Text != "" {
		text := strings.ToLower(*filter.Text)
		if !strings.Contains(strings.ToLower(issue.Title), text) {
			return false
		}
	}
	return true
}

func copyIssue(issue *models.Issue) *models.Issue {
	cp := *issue
	// Deep copy slices
	if issue.Assignees != nil {
		cp.Assignees = make([]uuid.UUID, len(issue.Assignees))
		copy(cp.Assignees, issue.Assignees)
	}
	if issue.Labels != nil {
		cp.Labels = make([]uuid.UUID, len(issue.Labels))
		copy(cp.Labels, issue.Labels)
	}
	if issue.RestrictedViewers != nil {
		cp.RestrictedViewers = make([]uuid.UUID, len(issue.RestrictedViewers))
		copy(cp.RestrictedViewers, issue.RestrictedViewers)
	}
	if issue.Checklist != nil {
		cp.Checklist = make([]models.ChecklistItem, len(issue.Checklist))
		copy(cp.Checklist, issue.Checklist)
	}
	return &cp
}
