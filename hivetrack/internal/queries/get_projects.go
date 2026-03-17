package queries

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/authentication"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type GetProjectsQuery struct{}

type ProjectSummary struct {
	ID          uuid.UUID
	Slug        string
	Name        string
	Description *string
	Archetype   models.ProjectArchetype
	Archived    bool
}

type GetProjectsResult struct {
	Projects []ProjectSummary
}

func HandleGetProjects(ctx context.Context, _ GetProjectsQuery) (*GetProjectsResult, error) {
	db := repositories.GetDbContext(ctx)
	actor := authentication.MustGetCurrentUser(ctx)

	filter := repositories.NewProjectFilter()
	if !actor.IsAdmin {
		filter = filter.ForMember(actor.ID)
	} else {
		filter = filter.AsAdmin()
	}

	projects, err := db.Projects().List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}

	summaries := make([]ProjectSummary, 0, len(projects))
	for _, p := range projects {
		summaries = append(summaries, ProjectSummary{
			ID:          p.ID,
			Slug:        p.Slug,
			Name:        p.Name,
			Description: p.Description,
			Archetype:   p.Archetype,
			Archived:    p.Archived,
		})
	}

	return &GetProjectsResult{Projects: summaries}, nil
}
