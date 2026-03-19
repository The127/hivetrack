package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/The127/mediatr"
	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/authentication"
	"github.com/the127/hivetrack/internal/events"
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
	OwnerID       *uuid.UUID
	ClearOwnerID  bool
	CancelReason  *string
	Refined       *bool
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

	oldStatus := issue.GetStatus()

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
	if cmd.ClearOwnerID {
		issue.SetOwnerID(nil)
	} else if cmd.OwnerID != nil {
		issue.SetOwnerID(cmd.OwnerID)
	}
	if cmd.CancelReason != nil {
		issue.SetCancelReason(cmd.CancelReason)
	}
	if cmd.Refined != nil {
		issue.SetRefined(*cmd.Refined)
	}

	if cmd.Status != nil && oldStatus == models.IssueStatusTodo && *cmd.Status == models.IssueStatusInProgress {
		if m, ok := getMediatorFromContext(ctx); ok {
			actor, _ := authentication.GetCurrentUser(ctx)
			if err := mediatr.SendEvent(ctx, m, events.IssueStatusChangedEvent{
				Issue:     issue,
				OldStatus: oldStatus,
				NewStatus: *cmd.Status,
				ActorID:   actor.ID,
			}); err != nil {
				return nil, fmt.Errorf("auto-assign on status change: %w", err)
			}
		}
	}

	issue.SetUpdatedAt(time.Now())

	db.Issues().Update(issue)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving issue: %w", err)
	}

	// Record status transition for burndown tracking
	if cmd.Status != nil && *cmd.Status != oldStatus {
		if err := db.IssueStatusLog().Insert(ctx, issue.GetId(), string(*cmd.Status), time.Now()); err != nil {
			return nil, fmt.Errorf("logging issue status: %w", err)
		}
	}

	return &UpdateIssueResult{}, nil
}
