package queries

import (
	"context"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/repositories"
)

// LabelInfo is a lightweight label representation for embedding in query responses.
type LabelInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}

// resolveLabels batch-fetches label details for a set of UUIDs and returns them as LabelInfo slices.
// Unknown IDs are silently skipped.
func resolveLabels(ctx context.Context, db repositories.DbContext, ids []uuid.UUID) ([]LabelInfo, error) {
	if len(ids) == 0 {
		return []LabelInfo{}, nil
	}

	result := make([]LabelInfo, 0, len(ids))
	for _, id := range ids {
		label, err := db.Labels().GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if label == nil {
			continue
		}
		result = append(result, LabelInfo{
			ID:    label.GetId(),
			Name:  label.GetName(),
			Color: label.GetColor(),
		})
	}
	return result, nil
}
