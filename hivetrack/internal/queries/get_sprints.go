package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type GetSprintsQuery struct {
	ProjectID uuid.UUID
}

type SprintSummary struct {
	ID        uuid.UUID
	Name      string
	Goal      *string
	StartDate time.Time
	EndDate   time.Time
	Status    models.SprintStatus
}

type GetSprintsResult struct {
	Sprints []SprintSummary
}

func HandleGetSprints(ctx context.Context, q GetSprintsQuery) (*GetSprintsResult, error) {
	db := repositories.GetDbContext(ctx)

	sprints, err := db.Sprints().List(ctx, q.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("listing sprints: %w", err)
	}

	summaries := make([]SprintSummary, 0, len(sprints))
	for _, s := range sprints {
		summaries = append(summaries, SprintSummary{
			ID:        s.GetId(),
			Name:      s.GetName(),
			Goal:      s.GetGoal(),
			StartDate: s.GetStartDate(),
			EndDate:   s.GetEndDate(),
			Status:    s.GetStatus(),
		})
	}

	return &GetSprintsResult{Sprints: summaries}, nil
}
