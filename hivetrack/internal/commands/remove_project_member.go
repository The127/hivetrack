package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type RemoveProjectMemberCommand struct {
	ProjectSlug string
	UserID      uuid.UUID
}

type RemoveProjectMemberResult struct{}

func HandleRemoveProjectMember(ctx context.Context, cmd RemoveProjectMemberCommand) (*RemoveProjectMemberResult, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetBySlug(ctx, cmd.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, models.ErrNotFound
	}

	// Verify membership exists
	existing, err := db.Projects().GetMember(ctx, project.GetId(), cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("checking membership: %w", err)
	}
	if existing == nil {
		return nil, models.ErrNotFound
	}

	if err := db.Projects().RemoveMember(ctx, project.GetId(), cmd.UserID); err != nil {
		return nil, fmt.Errorf("removing member: %w", err)
	}

	return &RemoveProjectMemberResult{}, nil
}
