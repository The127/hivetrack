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

		applyCommonFieldPatch(issue, issueFieldPatch{
			Status:        cmd.Status,
			Priority:      cmd.Priority,
			Estimate:      cmd.Estimate,
			AssigneeIDs:   cmd.AssigneeIDs,
			LabelIDs:      cmd.LabelIDs,
			SprintID:      cmd.SprintID,
			ClearSprintID: cmd.ClearSprintID,
			MilestoneID:   cmd.MilestoneID,
			OnHold:        cmd.OnHold,
			HoldReason:    cmd.HoldReason,
			HoldNote:      cmd.HoldNote,
		})

		issue.SetUpdatedAt(now)
		db.Issues().Update(issue)
		updated++
	}

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving batch update: %w", err)
	}

	return &BatchUpdateIssuesResult{Updated: updated}, nil
}
