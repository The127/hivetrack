package inmemory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
)

type LabelRepository struct {
	byID map[uuid.UUID]*models.Label
}

func NewLabelRepository() *LabelRepository {
	return &LabelRepository{
		byID: make(map[uuid.UUID]*models.Label),
	}
}

func (r *LabelRepository) Insert(_ context.Context, label *models.Label) error {
	cp := *label
	r.byID[label.ID] = &cp
	return nil
}

func (r *LabelRepository) Update(_ context.Context, label *models.Label) error {
	if _, ok := r.byID[label.ID]; !ok {
		return fmt.Errorf("label %s not found: %w", label.ID, models.ErrNotFound)
	}
	cp := *label
	r.byID[label.ID] = &cp
	return nil
}

func (r *LabelRepository) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := r.byID[id]; !ok {
		return fmt.Errorf("label %s not found: %w", id, models.ErrNotFound)
	}
	delete(r.byID, id)
	return nil
}

func (r *LabelRepository) GetByID(_ context.Context, id uuid.UUID) (*models.Label, error) {
	l, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	cp := *l
	return &cp, nil
}

func (r *LabelRepository) List(_ context.Context, projectID uuid.UUID) ([]*models.Label, error) {
	var result []*models.Label
	for _, l := range r.byID {
		if l.ProjectID == projectID {
			cp := *l
			result = append(result, &cp)
		}
	}
	return result, nil
}
