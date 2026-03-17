package commands

import (
	"context"
	"fmt"
	"time"

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

	project := &models.Project{
		ID:          uuid.New(),
		Slug:        cmd.Slug,
		Name:        cmd.Name,
		Archetype:   cmd.Archetype,
		Description: cmd.Description,
		CreatedBy:   actor.ID,
		CreatedAt:   time.Now(),
	}

	if err := db.Projects().Insert(ctx, project); err != nil {
		return nil, fmt.Errorf("inserting project: %w", err)
	}

	member := &models.ProjectMember{
		ProjectID: project.ID,
		UserID:    actor.ID,
		Role:      models.ProjectRoleAdmin,
	}
	if err := db.Projects().AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("adding creator as admin: %w", err)
	}

	if err := db.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing: %w", err)
	}

	return &CreateProjectResult{
		ID:   project.ID,
		Slug: project.Slug,
	}, nil
}
