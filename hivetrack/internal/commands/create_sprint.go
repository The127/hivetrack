package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type CreateSprintCommand struct {
	ProjectSlug string
	Name        string
	Goal        *string
	StartDate   *time.Time
	EndDate     *time.Time
}

type CreateSprintResult struct {
	ID uuid.UUID `json:"id"`
}

func HandleCreateSprint(ctx context.Context, cmd CreateSprintCommand) (*CreateSprintResult, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetBySlug(ctx, cmd.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %q: %w", cmd.ProjectSlug, models.ErrNotFound)
	}

	var startDate, endDate time.Time
	if cmd.StartDate != nil {
		startDate = *cmd.StartDate
	}
	if cmd.EndDate != nil {
		endDate = *cmd.EndDate
	}
	sprint := models.NewSprint(project.GetId(), cmd.Name, cmd.Goal, startDate, endDate, models.SprintStatusPlanning)
	db.Sprints().Insert(sprint)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving sprint: %w", err)
	}

	return &CreateSprintResult{ID: sprint.GetId()}, nil
}
