package queries

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/repositories"
)

type GetLabelsQuery struct {
	ProjectID uuid.UUID
}

type LabelSummary struct {
	ID    uuid.UUID
	Name  string
	Color string
}

type GetLabelsResult struct {
	Labels []LabelSummary
}

func HandleGetLabels(ctx context.Context, q GetLabelsQuery) (*GetLabelsResult, error) {
	db := repositories.GetDbContext(ctx)

	labels, err := db.Labels().List(ctx, q.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("listing labels: %w", err)
	}

	summaries := make([]LabelSummary, 0, len(labels))
	for _, l := range labels {
		summaries = append(summaries, LabelSummary{
			ID:    l.GetId(),
			Name:  l.GetName(),
			Color: l.GetColor(),
		})
	}

	return &GetLabelsResult{Labels: summaries}, nil
}
