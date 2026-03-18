package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type AddProjectMemberCommand struct {
	ProjectSlug string
	UserID      uuid.UUID
	Role        models.ProjectRole
}

type AddProjectMemberResult struct {
	ProjectID uuid.UUID `json:"project_id"`
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
}

func HandleAddProjectMember(ctx context.Context, cmd AddProjectMemberCommand) (*AddProjectMemberResult, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetBySlug(ctx, cmd.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, models.ErrNotFound
	}

	// Verify the target user exists
	user, err := db.Users().GetByID(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}
	if user == nil {
		return nil, models.ErrNotFound
	}

	// Check if already a member
	existing, err := db.Projects().GetMember(ctx, project.GetId(), cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("checking existing membership: %w", err)
	}
	if existing != nil {
		return nil, models.ErrConflict
	}

	member := &models.ProjectMember{
		ProjectID: project.GetId(),
		UserID:    cmd.UserID,
		Role:      cmd.Role,
	}
	if err := db.Projects().AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("adding member: %w", err)
	}

	return &AddProjectMemberResult{
		ProjectID: project.GetId(),
		UserID:    cmd.UserID,
		Role:      string(cmd.Role),
	}, nil
}
