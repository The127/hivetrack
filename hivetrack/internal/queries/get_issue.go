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
	ID          uuid.UUID
	ProjectID   uuid.UUID
	Number      int
	Type        models.IssueType
	Title       string
	Description *string
	Status      models.IssueStatus
	Priority    models.IssuePriority
	Estimate    models.IssueEstimate
	Triaged     bool
	Visibility  models.IssueVisibility
	OnHold      bool
	HoldReason  *models.HoldReason
	HoldNote    *string
	HoldSince   *time.Time
	Assignees   []uuid.UUID
	Labels      []uuid.UUID
	SprintID    *uuid.UUID
	MilestoneID *uuid.UUID
	ParentID    *uuid.UUID
	ReporterID  *uuid.UUID
	Checklist   []models.ChecklistItem
	CreatedAt   time.Time
	UpdatedAt   time.Time
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

	return &IssueDetail{
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
	}, nil
}
