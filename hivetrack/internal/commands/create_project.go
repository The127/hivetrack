package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/authentication"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type CreateProjectCommand struct {
	Slug        string
	Name        string
	Archetype   models.ProjectArchetype
	Description *string
}

type CreateProjectResult struct {
	ID   uuid.UUID `json:"id"`
	Slug string    `json:"slug"`
}

func HandleCreateProject(ctx context.Context, cmd CreateProjectCommand) (*CreateProjectResult, error) {
	db := repositories.GetDbContext(ctx)
	actor := authentication.MustGetCurrentUser(ctx)

	project := models.NewProject(actor.ID, cmd.Slug, cmd.Name, cmd.Archetype)
	if cmd.Description != nil {
		project.SetDescription(cmd.Description)
	}

	db.Projects().Insert(project)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving project: %w", err)
	}

	member := &models.ProjectMember{
		ProjectID: project.GetId(),
		UserID:    actor.ID,
		Role:      models.ProjectRoleAdmin,
	}
	if err := db.Projects().AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("adding creator as admin: %w", err)
	}

	return &CreateProjectResult{
		ID:   project.GetId(),
		Slug: project.GetSlug(),
	}, nil
}
