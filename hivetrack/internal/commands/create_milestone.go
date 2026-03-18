package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type CreateMilestoneCommand struct {
	ProjectSlug string
	Title       string
	Description *string
	TargetDate  *time.Time
}

type CreateMilestoneResult struct {
	ID uuid.UUID
}

func HandleCreateMilestone(ctx context.Context, cmd CreateMilestoneCommand) (*CreateMilestoneResult, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetBySlug(ctx, cmd.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %q: %w", cmd.ProjectSlug, models.ErrNotFound)
	}

	milestone := models.NewMilestone(project.GetId(), cmd.Title, cmd.Description, cmd.TargetDate)
	db.Milestones().Insert(milestone)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving milestone: %w", err)
	}

	return &CreateMilestoneResult{ID: milestone.GetId()}, nil
}
