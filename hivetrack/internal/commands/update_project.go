package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type UpdateProjectCommand struct {
	ID          uuid.UUID
	Name        *string
	Description *string
	Archived    *bool
}

type UpdateProjectResult struct{}

func HandleUpdateProject(ctx context.Context, cmd UpdateProjectCommand) (*UpdateProjectResult, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %s: %w", cmd.ID, models.ErrNotFound)
	}

	if cmd.Name != nil {
		project.SetName(*cmd.Name)
	}
	if cmd.Description != nil {
		project.SetDescription(cmd.Description)
	}
	if cmd.Archived != nil {
		project.SetArchived(*cmd.Archived)
	}

	db.Projects().Update(project)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving project: %w", err)
	}

	return &UpdateProjectResult{}, nil
}
