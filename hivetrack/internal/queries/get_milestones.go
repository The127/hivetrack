package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/repositories"
)

type GetMilestonesQuery struct {
	ProjectID uuid.UUID
}

type MilestoneSummary struct {
	ID          uuid.UUID
	Title       string
	Description *string
	TargetDate  *time.Time
	ClosedAt    *time.Time
}

type GetMilestonesResult struct {
	Milestones []MilestoneSummary
}

func HandleGetMilestones(ctx context.Context, q GetMilestonesQuery) (*GetMilestonesResult, error) {
	db := repositories.GetDbContext(ctx)

	milestones, err := db.Milestones().List(ctx, q.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("listing milestones: %w", err)
	}

	summaries := make([]MilestoneSummary, 0, len(milestones))
	for _, m := range milestones {
		summaries = append(summaries, MilestoneSummary{
			ID:          m.GetId(),
			Title:       m.GetTitle(),
			Description: m.GetDescription(),
			TargetDate:  m.GetTargetDate(),
			ClosedAt:    m.GetClosedAt(),
		})
	}

	return &GetMilestonesResult{Milestones: summaries}, nil
}
