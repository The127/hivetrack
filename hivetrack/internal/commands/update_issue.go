package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type UpdateIssueCommand struct {
	IssueID     uuid.UUID
	Title       *string
	Description *string
	Status      *models.IssueStatus
	Priority    *models.IssuePriority
	Estimate    *models.IssueEstimate
	AssigneeIDs []uuid.UUID
	LabelIDs    []uuid.UUID
	SprintID    *uuid.UUID
	MilestoneID *uuid.UUID
	OnHold      *bool
	HoldReason  *models.HoldReason
	HoldNote    *string
	Visibility  *models.IssueVisibility
}

type UpdateIssueResult struct{}

func HandleUpdateIssue(ctx context.Context, cmd UpdateIssueCommand) (*UpdateIssueResult, error) {
	db := repositories.GetDbContext(ctx)

	issue, err := db.Issues().GetByID(ctx, cmd.IssueID)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue %s: %w", cmd.IssueID, models.ErrNotFound)
	}

	if cmd.Title != nil {
		issue.Title = *cmd.Title
	}
	if cmd.Description != nil {
		issue.Description = cmd.Description
	}
	if cmd.Status != nil {
		issue.Status = *cmd.Status
	}
	if cmd.Priority != nil {
		issue.Priority = *cmd.Priority
	}
	if cmd.Estimate != nil {
		issue.Estimate = *cmd.Estimate
	}
	if cmd.AssigneeIDs != nil {
		issue.Assignees = cmd.AssigneeIDs
	}
	if cmd.LabelIDs != nil {
		issue.Labels = cmd.LabelIDs
	}
	if cmd.SprintID != nil {
		issue.SprintID = cmd.SprintID
	}
	if cmd.MilestoneID != nil {
		issue.MilestoneID = cmd.MilestoneID
	}
	if cmd.OnHold != nil {
		issue.OnHold = *cmd.OnHold
		if *cmd.OnHold {
			now := time.Now()
			issue.HoldSince = &now
			issue.HoldReason = cmd.HoldReason
			issue.HoldNote = cmd.HoldNote
		} else {
			issue.HoldSince = nil
			issue.HoldReason = nil
			issue.HoldNote = nil
		}
	}
	if cmd.Visibility != nil {
		issue.Visibility = *cmd.Visibility
	}

	issue.UpdatedAt = time.Now()

	if err := db.Issues().Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("updating issue: %w", err)
	}

	if err := db.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing: %w", err)
	}

	return &UpdateIssueResult{}, nil
}
