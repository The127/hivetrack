package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type BatchUpdateIssuesCommand struct {
	ProjectID     uuid.UUID
	IssueNumbers  []int
	Status        *models.IssueStatus
	Priority      *models.IssuePriority
	Estimate      *models.IssueEstimate
	AssigneeIDs   []uuid.UUID
	LabelIDs      []uuid.UUID
	SprintID      *uuid.UUID
	ClearSprintID bool
	MilestoneID   *uuid.UUID
	OnHold        *bool
	HoldReason    *models.HoldReason
	HoldNote      *string
}

type BatchUpdateIssuesResult struct {
	Updated int
}

func HandleBatchUpdateIssues(ctx context.Context, cmd BatchUpdateIssuesCommand) (*BatchUpdateIssuesResult, error) {
	db := repositories.GetDbContext(ctx)

	if len(cmd.IssueNumbers) == 0 {
		return nil, fmt.Errorf("no issue numbers provided: %w", models.ErrBadRequest)
	}

	now := time.Now()
	updated := 0

	for _, number := range cmd.IssueNumbers {
		issue, err := db.Issues().GetByNumber(ctx, cmd.ProjectID, number)
		if err != nil {
			return nil, fmt.Errorf("getting issue #%d: %w", number, err)
		}
		if issue == nil {
			return nil, fmt.Errorf("issue #%d: %w", number, models.ErrNotFound)
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
		if cmd.OnHold != nil {
			if *cmd.OnHold {
				issue.SetHold(true, cmd.HoldReason, &now, cmd.HoldNote)
			} else {
				issue.SetHold(false, nil, nil, nil)
			}
		}

		issue.SetUpdatedAt(now)
		db.Issues().Update(issue)
		updated++
	}

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving batch update: %w", err)
	}

	return &BatchUpdateIssuesResult{Updated: updated}, nil
}
