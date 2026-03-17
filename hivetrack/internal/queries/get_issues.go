package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type GetIssuesQuery struct {
	ProjectSlug string
	Status      *models.IssueStatus
	Priority    *models.IssuePriority
	SprintID    *uuid.UUID
	AssigneeID  *uuid.UUID
	Triaged     *bool
	Text        *string
	Limit       int
	Offset      int
}

type IssueSummary struct {
	ID          uuid.UUID            `json:"id"`
	Number      int                  `json:"number"`
	Type        models.IssueType     `json:"type"`
	Title       string               `json:"title"`
	Status      models.IssueStatus   `json:"status"`
	Priority    models.IssuePriority `json:"priority"`
	Estimate    models.IssueEstimate `json:"estimate"`
	Triaged     bool                 `json:"triaged"`
	Assignees   []uuid.UUID          `json:"assignees"`
	Labels      []uuid.UUID          `json:"labels"`
	SprintID    *uuid.UUID           `json:"sprint_id,omitempty"`
	MilestoneID *uuid.UUID           `json:"milestone_id,omitempty"`
	OnHold      bool                 `json:"on_hold"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type GetIssuesResult struct {
	Items  []IssueSummary `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

func HandleGetIssues(ctx context.Context, q GetIssuesQuery) (*GetIssuesResult, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetBySlug(ctx, q.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %q: %w", q.ProjectSlug, models.ErrNotFound)
	}

	filter := repositories.NewIssueFilter().ByProjectID(project.GetId())
	if q.Status != nil {
		filter = filter.ByStatus(*q.Status)
	}
	if q.Priority != nil {
		filter = filter.ByPriority(*q.Priority)
	}
	if q.SprintID != nil {
		filter = filter.BySprintID(*q.SprintID)
	}
	if q.AssigneeID != nil {
		filter = filter.ByAssigneeID(*q.AssigneeID)
	}
	if q.Triaged != nil {
		filter = filter.WithTriaged(*q.Triaged)
	}
	if q.Text != nil {
		filter = filter.WithText(*q.Text)
	}
	if q.Limit > 0 || q.Offset > 0 {
		filter = filter.WithPagination(q.Limit, q.Offset)
	}

	issues, total, err := db.Issues().List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing issues: %w", err)
	}

	items := make([]IssueSummary, 0, len(issues))
	for _, i := range issues {
		items = append(items, IssueSummary{
			ID:          i.GetId(),
			Number:      i.GetNumber(),
			Type:        i.GetType(),
			Title:       i.GetTitle(),
			Status:      i.GetStatus(),
			Priority:    i.GetPriority(),
			Estimate:    i.GetEstimate(),
			Triaged:     i.GetTriaged(),
			Assignees:   i.GetAssignees(),
			Labels:      i.GetLabels(),
			SprintID:    i.GetSprintID(),
			MilestoneID: i.GetMilestoneID(),
			OnHold:      i.GetOnHold(),
			CreatedAt:   i.GetCreatedAt(),
			UpdatedAt:   i.GetUpdatedAt(),
		})
	}

	return &GetIssuesResult{
		Items:  items,
		Total:  total,
		Limit:  q.Limit,
		Offset: q.Offset,
	}, nil
}
