package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type DeleteSprintCommand struct {
	SprintID uuid.UUID
}

type DeleteSprintResult struct{}

func HandleDeleteSprint(ctx context.Context, cmd DeleteSprintCommand) (*DeleteSprintResult, error) {
	db := repositories.GetDbContext(ctx)

	sprint, err := db.Sprints().GetByID(ctx, cmd.SprintID)
	if err != nil {
		return nil, fmt.Errorf("getting sprint: %w", err)
	}
	if sprint == nil {
		return nil, fmt.Errorf("sprint %s: %w", cmd.SprintID, models.ErrNotFound)
	}

	// Move all sprint issues to the backlog.
	filter := repositories.NewIssueFilter().BySprintID(sprint.GetId())
	issues, _, err := db.Issues().List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing sprint issues: %w", err)
	}
	for _, issue := range issues {
		issue.SetSprintID(nil)
		issue.SetUpdatedAt(time.Now())
		db.Issues().Update(issue)
	}

	db.Sprints().Delete(sprint)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("deleting sprint: %w", err)
	}

	return &DeleteSprintResult{}, nil
}
