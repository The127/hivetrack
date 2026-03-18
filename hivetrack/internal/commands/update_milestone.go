package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type UpdateMilestoneCommand struct {
	MilestoneID uuid.UUID
	Title       *string
	Description *string
	TargetDate  *time.Time
	Close       *bool
}

type UpdateMilestoneResult struct{}

func HandleUpdateMilestone(ctx context.Context, cmd UpdateMilestoneCommand) (*UpdateMilestoneResult, error) {
	db := repositories.GetDbContext(ctx)

	milestone, err := db.Milestones().GetByID(ctx, cmd.MilestoneID)
	if err != nil {
		return nil, fmt.Errorf("getting milestone: %w", err)
	}
	if milestone == nil {
		return nil, fmt.Errorf("milestone %s: %w", cmd.MilestoneID, models.ErrNotFound)
	}

	if cmd.Title != nil {
		milestone.SetTitle(*cmd.Title)
	}
	if cmd.Description != nil {
		milestone.SetDescription(cmd.Description)
	}
	if cmd.TargetDate != nil {
		milestone.SetTargetDate(cmd.TargetDate)
	}
	if cmd.Close != nil {
		if *cmd.Close {
			now := time.Now()
			milestone.SetClosedAt(&now)
		} else {
			milestone.SetClosedAt(nil)
		}
	}

	db.Milestones().Update(milestone)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving milestone: %w", err)
	}

	return &UpdateMilestoneResult{}, nil
}
