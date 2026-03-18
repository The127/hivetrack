package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type CreateLabelCommand struct {
	ProjectSlug string
	Name        string
	Color       string
}

type CreateLabelResult struct {
	ID uuid.UUID
}

func HandleCreateLabel(ctx context.Context, cmd CreateLabelCommand) (*CreateLabelResult, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetBySlug(ctx, cmd.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %q: %w", cmd.ProjectSlug, models.ErrNotFound)
	}

	label := models.NewLabel(project.GetId(), cmd.Name, cmd.Color)
	db.Labels().Insert(label)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving label: %w", err)
	}

	return &CreateLabelResult{ID: label.GetId()}, nil
}
