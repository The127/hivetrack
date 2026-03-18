package queries

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type GetLabelsQuery struct {
	ProjectSlug string
}

type LabelSummary struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}

type GetLabelsResult struct {
	Labels []LabelSummary `json:"labels"`
}

func HandleGetLabels(ctx context.Context, q GetLabelsQuery) (*GetLabelsResult, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetBySlug(ctx, q.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %q: %w", q.ProjectSlug, models.ErrNotFound)
	}

	labels, err := db.Labels().List(ctx, project.GetId())
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
