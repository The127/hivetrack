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
		project.Name = *cmd.Name
	}
	if cmd.Description != nil {
		project.Description = cmd.Description
	}
	if cmd.Archived != nil {
		project.Archived = *cmd.Archived
	}

	if err := db.Projects().Update(ctx, project); err != nil {
		return nil, fmt.Errorf("updating project: %w", err)
	}

	if err := db.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing: %w", err)
	}

	return &UpdateProjectResult{}, nil
}
