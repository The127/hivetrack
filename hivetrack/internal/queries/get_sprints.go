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
	ProjectSlug string
}

type SprintSummary struct {
	ID        uuid.UUID           `json:"id"`
	Name      string              `json:"name"`
	Goal      *string             `json:"goal,omitempty"`
	StartDate *time.Time          `json:"start_date,omitempty"`
	EndDate   *time.Time          `json:"end_date,omitempty"`
	Status    models.SprintStatus `json:"status"`
}

type GetSprintsResult struct {
	Sprints []SprintSummary `json:"sprints"`
}

func HandleGetSprints(ctx context.Context, q GetSprintsQuery) (*GetSprintsResult, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetBySlug(ctx, q.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %q: %w", q.ProjectSlug, models.ErrNotFound)
	}

	sprints, err := db.Sprints().List(ctx, project.GetId())
	if err != nil {
		return nil, fmt.Errorf("listing sprints: %w", err)
	}

	summaries := make([]SprintSummary, 0, len(sprints))
	for _, s := range sprints {
		sum := SprintSummary{
			ID:     s.GetId(),
			Name:   s.GetName(),
			Goal:   s.GetGoal(),
			Status: s.GetStatus(),
		}
		if !s.GetStartDate().IsZero() {
			t := s.GetStartDate()
			sum.StartDate = &t
		}
		if !s.GetEndDate().IsZero() {
			t := s.GetEndDate()
			sum.EndDate = &t
		}
		summaries = append(summaries, sum)
	}

	return &GetSprintsResult{Sprints: summaries}, nil
}
