package queries

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type GetProjectQuery struct {
	Slug string
}

type ProjectMemberInfo struct {
	UserID uuid.UUID
	Role   models.ProjectRole
}

type GetProjectResult struct {
	ID          uuid.UUID
	Slug        string
	Name        string
	Description *string
	Archetype   models.ProjectArchetype
	Archived    bool
	CreatedBy   uuid.UUID
	Members     []ProjectMemberInfo
}

func HandleGetProject(ctx context.Context, q GetProjectQuery) (*GetProjectResult, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetBySlug(ctx, q.Slug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, nil
	}

	members, err := db.Projects().ListMembers(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}

	memberInfos := make([]ProjectMemberInfo, 0, len(members))
	for _, m := range members {
		memberInfos = append(memberInfos, ProjectMemberInfo{
			UserID: m.UserID,
			Role:   m.Role,
		})
	}

	return &GetProjectResult{
		ID:          project.ID,
		Slug:        project.Slug,
		Name:        project.Name,
		Description: project.Description,
		Archetype:   project.Archetype,
		Archived:    project.Archived,
		CreatedBy:   project.CreatedBy,
		Members:     memberInfos,
	}, nil
}
