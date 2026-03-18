package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type UpdateLabelCommand struct {
	LabelID uuid.UUID
	Name    *string
	Color   *string
}

type UpdateLabelResult struct{}

func HandleUpdateLabel(ctx context.Context, cmd UpdateLabelCommand) (*UpdateLabelResult, error) {
	db := repositories.GetDbContext(ctx)

	label, err := db.Labels().GetByID(ctx, cmd.LabelID)
	if err != nil {
		return nil, fmt.Errorf("getting label: %w", err)
	}
	if label == nil {
		return nil, fmt.Errorf("label %s: %w", cmd.LabelID, models.ErrNotFound)
	}

	if cmd.Name != nil {
		label.SetName(*cmd.Name)
	}
	if cmd.Color != nil {
		label.SetColor(*cmd.Color)
	}

	db.Labels().Update(label)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving label: %w", err)
	}

	return &UpdateLabelResult{}, nil
}
