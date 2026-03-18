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
	UserID      uuid.UUID          `json:"user_id"`
	Role        models.ProjectRole `json:"role"`
	DisplayName string             `json:"display_name"`
	AvatarURL   *string            `json:"avatar_url,omitempty"`
}

type GetProjectResult struct {
	ID          uuid.UUID               `json:"id"`
	Slug        string                  `json:"slug"`
	Name        string                  `json:"name"`
	Description *string                 `json:"description,omitempty"`
	Archetype   models.ProjectArchetype `json:"archetype"`
	Archived    bool                    `json:"archived"`
	CreatedBy   uuid.UUID               `json:"created_by"`
	Members     []ProjectMemberInfo     `json:"members"`
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

	members, err := db.Projects().ListMembers(ctx, project.GetId())
	if err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}

	memberInfos := make([]ProjectMemberInfo, 0, len(members))
	for _, m := range members {
		info := ProjectMemberInfo{
			UserID: m.UserID,
			Role:   m.Role,
		}
		user, err := db.Users().GetByID(ctx, m.UserID)
		if err != nil {
			return nil, fmt.Errorf("getting user %s: %w", m.UserID, err)
		}
		if user != nil {
			info.DisplayName = user.GetDisplayName()
			info.AvatarURL = user.GetAvatarURL()
		}
		memberInfos = append(memberInfos, info)
	}

	return &GetProjectResult{
		ID:          project.GetId(),
		Slug:        project.GetSlug(),
		Name:        project.GetName(),
		Description: project.GetDescription(),
		Archetype:   project.GetArchetype(),
		Archived:    project.GetArchived(),
		CreatedBy:   project.GetCreatedBy(),
		Members:     memberInfos,
	}, nil
}
