package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/authentication"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type CreateIssueCommand struct {
	ProjectSlug string
	Title       string
	Type        models.IssueType
	Priority    *models.IssuePriority
	Estimate    *models.IssueEstimate
	Status      *models.IssueStatus
	Description *string
	AssigneeIDs []uuid.UUID
	LabelIDs    []uuid.UUID
	SprintID    *uuid.UUID
	MilestoneID *uuid.UUID
}

type CreateIssueResult struct {
	ID     uuid.UUID
	Number int
}

func HandleCreateIssue(ctx context.Context, cmd CreateIssueCommand) (*CreateIssueResult, error) {
	db := repositories.GetDbContext(ctx)
	actor := authentication.MustGetCurrentUser(ctx)

	project, err := db.Projects().GetBySlug(ctx, cmd.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %q: %w", cmd.ProjectSlug, models.ErrNotFound)
	}

	number, err := db.Projects().NextIssueNumber(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("getting next issue number: %w", err)
	}

	// Determine default status
	status := models.IssueStatus("")
	if cmd.Status != nil {
		status = *cmd.Status
	} else {
		switch project.Archetype {
		case models.ProjectArchetypeSoftware:
			status = models.IssueStatusTodo
		case models.ProjectArchetypeSupport:
			status = models.IssueStatusOpen
		}
	}

	// Quick-capture: triaged=false only if no status, sprint, or milestone given
	triaged := cmd.Status != nil || cmd.SprintID != nil || cmd.MilestoneID != nil

	priority := models.IssuePriorityNone
	if cmd.Priority != nil {
		priority = *cmd.Priority
	}
	estimate := models.IssueEstimateNone
	if cmd.Estimate != nil {
		estimate = *cmd.Estimate
	}

	reporterID := actor.ID
	now := time.Now()

	issue := &models.Issue{
		ID:          uuid.New(),
		ProjectID:   project.ID,
		Number:      number,
		Type:        cmd.Type,
		Title:       cmd.Title,
		Description: cmd.Description,
		Status:      status,
		Priority:    priority,
		Estimate:    estimate,
		ReporterID:  &reporterID,
		SprintID:    cmd.SprintID,
		MilestoneID: cmd.MilestoneID,
		Triaged:     triaged,
		Visibility:  models.IssueVisibilityNormal,
		Assignees:   cmd.AssigneeIDs,
		Labels:      cmd.LabelIDs,
		Checklist:   []models.ChecklistItem{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := db.Issues().Insert(ctx, issue); err != nil {
		return nil, fmt.Errorf("inserting issue: %w", err)
	}

	if err := db.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing: %w", err)
	}

	return &CreateIssueResult{
		ID:     issue.ID,
		Number: issue.Number,
	}, nil
}
