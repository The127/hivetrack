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
	IssueID       uuid.UUID
	Title         *string
	Description   *string
	Status        *models.IssueStatus
	Priority      *models.IssuePriority
	Estimate      *models.IssueEstimate
	AssigneeIDs   []uuid.UUID
	LabelIDs      []uuid.UUID
	SprintID      *uuid.UUID
	ClearSprintID bool
	MilestoneID   *uuid.UUID
	ParentID      *uuid.UUID
	ClearParentID bool
	OnHold        *bool
	HoldReason    *models.HoldReason
	HoldNote      *string
	Visibility    *models.IssueVisibility
	Rank          *string
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
		issue.SetTitle(*cmd.Title)
	}
	if cmd.Description != nil {
		issue.SetDescription(cmd.Description)
	}
	if cmd.Status != nil {
		issue.SetStatus(*cmd.Status)
	}
	if cmd.Priority != nil {
		issue.SetPriority(*cmd.Priority)
	}
	if cmd.Estimate != nil {
		issue.SetEstimate(*cmd.Estimate)
	}
	if cmd.AssigneeIDs != nil {
		issue.SetAssignees(cmd.AssigneeIDs)
	}
	if cmd.LabelIDs != nil {
		issue.SetLabels(cmd.LabelIDs)
	}
	if cmd.ClearSprintID {
		issue.SetSprintID(nil)
	} else if cmd.SprintID != nil {
		issue.SetSprintID(cmd.SprintID)
	}
	if cmd.MilestoneID != nil {
		issue.SetMilestoneID(cmd.MilestoneID)
	}
	if cmd.ClearParentID {
		issue.SetParentID(nil)
	} else if cmd.ParentID != nil {
		if issue.GetType() != models.IssueTypeTask {
			return nil, fmt.Errorf("only tasks can have a parent: %w", models.ErrBadRequest)
		}
		parent, err := db.Issues().GetByID(ctx, *cmd.ParentID)
		if err != nil {
			return nil, fmt.Errorf("getting parent issue: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("parent issue %s: %w", cmd.ParentID, models.ErrNotFound)
		}
		if parent.GetType() != models.IssueTypeEpic {
			return nil, fmt.Errorf("parent must be an epic: %w", models.ErrBadRequest)
		}
		if parent.GetProjectID() != issue.GetProjectID() {
			return nil, fmt.Errorf("parent must be in the same project: %w", models.ErrBadRequest)
		}
		issue.SetParentID(cmd.ParentID)
	}
	if cmd.OnHold != nil {
		if *cmd.OnHold {
			now := time.Now()
			issue.SetHold(true, cmd.HoldReason, &now, cmd.HoldNote)
		} else {
			issue.SetHold(false, nil, nil, nil)
		}
	}
	if cmd.Visibility != nil {
		issue.SetVisibility(*cmd.Visibility)
	}
	if cmd.Rank != nil {
		issue.SetRank(cmd.Rank)
	}

	issue.SetUpdatedAt(time.Now())

	db.Issues().Update(issue)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving issue: %w", err)
	}

	return &UpdateIssueResult{}, nil
}
