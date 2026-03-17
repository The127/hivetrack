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

	issue, err := db.Issues().GetByNumber(ctx, project.ID, q.Number)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, nil
	}

	return &IssueDetail{
		ID:          issue.ID,
		ProjectID:   issue.ProjectID,
		Number:      issue.Number,
		Type:        issue.Type,
		Title:       issue.Title,
		Description: issue.Description,
		Status:      issue.Status,
		Priority:    issue.Priority,
		Estimate:    issue.Estimate,
		Triaged:     issue.Triaged,
		Visibility:  issue.Visibility,
		OnHold:      issue.OnHold,
		HoldReason:  issue.HoldReason,
		HoldNote:    issue.HoldNote,
		HoldSince:   issue.HoldSince,
		Assignees:   issue.Assignees,
		Labels:      issue.Labels,
		SprintID:    issue.SprintID,
		MilestoneID: issue.MilestoneID,
		ParentID:    issue.ParentID,
		ReporterID:  issue.ReporterID,
		Checklist:   issue.Checklist,
		CreatedAt:   issue.CreatedAt,
		UpdatedAt:   issue.UpdatedAt,
	}, nil
}
