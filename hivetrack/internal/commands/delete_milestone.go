package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type DeleteMilestoneCommand struct {
	MilestoneID uuid.UUID
}

type DeleteMilestoneResult struct{}

func HandleDeleteMilestone(ctx context.Context, cmd DeleteMilestoneCommand) (*DeleteMilestoneResult, error) {
	db := repositories.GetDbContext(ctx)

	milestone, err := db.Milestones().GetByID(ctx, cmd.MilestoneID)
	if err != nil {
		return nil, fmt.Errorf("getting milestone: %w", err)
	}
	if milestone == nil {
		return nil, fmt.Errorf("milestone %s: %w", cmd.MilestoneID, models.ErrNotFound)
	}

	db.Milestones().Delete(milestone)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("deleting milestone: %w", err)
	}

	return &DeleteMilestoneResult{}, nil
}
