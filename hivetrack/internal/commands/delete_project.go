package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type DeleteProjectCommand struct {
	ID uuid.UUID
}

type DeleteProjectResult struct{}

func HandleDeleteProject(ctx context.Context, cmd DeleteProjectCommand) (*DeleteProjectResult, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %s: %w", cmd.ID, models.ErrNotFound)
	}

	if err := db.Projects().Delete(ctx, cmd.ID); err != nil {
		return nil, fmt.Errorf("deleting project: %w", err)
	}

	if err := db.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing: %w", err)
	}

	return &DeleteProjectResult{}, nil
}
