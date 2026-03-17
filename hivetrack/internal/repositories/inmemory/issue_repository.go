package inmemory

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/change"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type IssueRepository struct {
	tracker            *change.Tracker
	byID               map[uuid.UUID]*models.Issue
	byProjectAndNumber map[string]*models.Issue // "projectID:number" -> issue
}

func NewIssueRepository(tracker *change.Tracker) *IssueRepository {
	return &IssueRepository{
		tracker:            tracker,
		byID:               make(map[uuid.UUID]*models.Issue),
		byProjectAndNumber: make(map[string]*models.Issue),
	}
}

func issueKey(projectID uuid.UUID, number int) string {
	return fmt.Sprintf("%s:%d", projectID, number)
}

func (r *IssueRepository) Insert(issue *models.Issue) {
	r.tracker.Add(change.NewEntry(0, issue, change.Added))
}

func (r *IssueRepository) Update(issue *models.Issue) {
	r.tracker.Add(change.NewEntry(0, issue, change.Updated))
}

func (r *IssueRepository) Delete(issue *models.Issue) {
	r.tracker.Add(change.NewEntry(0, issue, change.Deleted))
}

func (r *IssueRepository) GetByID(_ context.Context, id uuid.UUID) (*models.Issue, error) {
	issue, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	return issue, nil
}

func (r *IssueRepository) GetByNumber(_ context.Context, projectID uuid.UUID, number int) (*models.Issue, error) {
	issue, ok := r.byProjectAndNumber[issueKey(projectID, number)]
	if !ok {
		return nil, nil
	}
	return issue, nil
}

func (r *IssueRepository) List(_ context.Context, filter *repositories.IssueFilter) ([]*models.Issue, int, error) {
	var result []*models.Issue
	for _, issue := range r.byID {
		if !matchesIssueFilter(issue, filter) {
			continue
		}
		result = append(result, issue)
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
	if filter.ProjectID != nil && issue.GetProjectID() != *filter.ProjectID {
		return false
	}
	if filter.Status != nil && issue.GetStatus() != *filter.Status {
		return false
	}
	if filter.Priority != nil && issue.GetPriority() != *filter.Priority {
		return false
	}
	if filter.SprintID != nil {
		if issue.GetSprintID() == nil || *issue.GetSprintID() != *filter.SprintID {
			return false
		}
	}
	if filter.AssigneeID != nil {
		found := false
		for _, a := range issue.GetAssignees() {
			if a == *filter.AssigneeID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if filter.Triaged != nil && issue.GetTriaged() != *filter.Triaged {
		return false
	}
	if filter.Text != nil && *filter.Text != "" {
		text := strings.ToLower(*filter.Text)
		if !strings.Contains(strings.ToLower(issue.GetTitle()), text) {
			return false
		}
	}
	return true
}
