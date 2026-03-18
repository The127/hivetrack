package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type DeleteLabelCommand struct {
	LabelID uuid.UUID
}

type DeleteLabelResult struct{}

func HandleDeleteLabel(ctx context.Context, cmd DeleteLabelCommand) (*DeleteLabelResult, error) {
	db := repositories.GetDbContext(ctx)

	label, err := db.Labels().GetByID(ctx, cmd.LabelID)
	if err != nil {
		return nil, fmt.Errorf("getting label: %w", err)
	}
	if label == nil {
		return nil, fmt.Errorf("label %s: %w", cmd.LabelID, models.ErrNotFound)
	}

	db.Labels().Delete(label)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("deleting label: %w", err)
	}

	return &DeleteLabelResult{}, nil
}
