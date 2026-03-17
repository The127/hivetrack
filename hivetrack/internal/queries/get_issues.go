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
	ID          uuid.UUID
	Number      int
	Type        models.IssueType
	Title       string
	Status      models.IssueStatus
	Priority    models.IssuePriority
	Estimate    models.IssueEstimate
	Triaged     bool
	Assignees   []uuid.UUID
	Labels      []uuid.UUID
	SprintID    *uuid.UUID
	MilestoneID *uuid.UUID
	OnHold      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GetIssuesResult struct {
	Items  []IssueSummary
	Total  int
	Limit  int
	Offset int
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

	filter := repositories.NewIssueFilter().ByProjectID(project.ID)
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
			ID:          i.ID,
			Number:      i.Number,
			Type:        i.Type,
			Title:       i.Title,
			Status:      i.Status,
			Priority:    i.Priority,
			Estimate:    i.Estimate,
			Triaged:     i.Triaged,
			Assignees:   i.Assignees,
			Labels:      i.Labels,
			SprintID:    i.SprintID,
			MilestoneID: i.MilestoneID,
			OnHold:      i.OnHold,
			CreatedAt:   i.CreatedAt,
			UpdatedAt:   i.UpdatedAt,
		})
	}

	return &GetIssuesResult{
		Items:  items,
		Total:  total,
		Limit:  q.Limit,
		Offset: q.Offset,
	}, nil
}
