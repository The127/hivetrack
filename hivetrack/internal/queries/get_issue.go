package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type GetIssueQuery struct {
	ProjectSlug string
	Number      int
}

type IssueDetail struct {
	ID             uuid.UUID              `json:"id"`
	ProjectID      uuid.UUID              `json:"project_id"`
	Number         int                    `json:"number"`
	Type           models.IssueType       `json:"type"`
	Title          string                 `json:"title"`
	Description    *string                `json:"description,omitempty"`
	Status         models.IssueStatus     `json:"status"`
	Priority       models.IssuePriority   `json:"priority"`
	Estimate       models.IssueEstimate   `json:"estimate"`
	Triaged        bool                   `json:"triaged"`
	Visibility     models.IssueVisibility `json:"visibility"`
	OnHold         bool                   `json:"on_hold"`
	HoldReason     *models.HoldReason     `json:"hold_reason,omitempty"`
	HoldNote       *string                `json:"hold_note,omitempty"`
	HoldSince      *time.Time             `json:"hold_since,omitempty"`
	Assignees      []uuid.UUID            `json:"assignees"`
	Labels         []uuid.UUID            `json:"labels"`
	SprintID       *uuid.UUID             `json:"sprint_id,omitempty"`
	MilestoneID    *uuid.UUID             `json:"milestone_id,omitempty"`
	ParentID       *uuid.UUID             `json:"parent_id,omitempty"`
	ReporterID     *uuid.UUID             `json:"reporter_id,omitempty"`
	Checklist      []models.ChecklistItem `json:"checklist"`
	ChildCount     int                    `json:"child_count"`
	ChildDoneCount int                    `json:"child_done_count"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

func HandleGetIssue(ctx context.Context, q GetIssueQuery) (*IssueDetail, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetBySlug(ctx, q.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %q: %w", q.ProjectSlug, models.ErrNotFound)
	}

	issue, err := db.Issues().GetByNumber(ctx, project.GetId(), q.Number)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, nil
	}

	detail := &IssueDetail{
		ID:          issue.GetId(),
		ProjectID:   issue.GetProjectID(),
		Number:      issue.GetNumber(),
		Type:        issue.GetType(),
		Title:       issue.GetTitle(),
		Description: issue.GetDescription(),
		Status:      issue.GetStatus(),
		Priority:    issue.GetPriority(),
		Estimate:    issue.GetEstimate(),
		Triaged:     issue.GetTriaged(),
		Visibility:  issue.GetVisibility(),
		OnHold:      issue.GetOnHold(),
		HoldReason:  issue.GetHoldReason(),
		HoldNote:    issue.GetHoldNote(),
		HoldSince:   issue.GetHoldSince(),
		Assignees:   issue.GetAssignees(),
		Labels:      issue.GetLabels(),
		SprintID:    issue.GetSprintID(),
		MilestoneID: issue.GetMilestoneID(),
		ParentID:    issue.GetParentID(),
		ReporterID:  issue.GetReporterID(),
		Checklist:   issue.GetChecklist(),
		CreatedAt:   issue.GetCreatedAt(),
		UpdatedAt:   issue.GetUpdatedAt(),
	}

	if issue.GetType() == models.IssueTypeEpic {
		children, _, err := db.Issues().List(ctx, repositories.NewIssueFilter().ByParentID(issue.GetId()))
		if err != nil {
			return nil, fmt.Errorf("listing child issues: %w", err)
		}
		detail.ChildCount = len(children)
		for _, child := range children {
			if child.IsTerminal() {
				detail.ChildDoneCount++
			}
		}
	}

	return detail, nil
}
