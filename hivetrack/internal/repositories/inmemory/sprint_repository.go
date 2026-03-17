package inmemory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
)

type SprintRepository struct {
	byID map[uuid.UUID]*models.Sprint
}

func NewSprintRepository() *SprintRepository {
	return &SprintRepository{
		byID: make(map[uuid.UUID]*models.Sprint),
	}
}

func (r *SprintRepository) Insert(_ context.Context, sprint *models.Sprint) error {
	cp := *sprint
	r.byID[sprint.ID] = &cp
	return nil
}

func (r *SprintRepository) Update(_ context.Context, sprint *models.Sprint) error {
	if _, ok := r.byID[sprint.ID]; !ok {
		return fmt.Errorf("sprint %s not found: %w", sprint.ID, models.ErrNotFound)
	}
	cp := *sprint
	r.byID[sprint.ID] = &cp
	return nil
}

func (r *SprintRepository) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := r.byID[id]; !ok {
		return fmt.Errorf("sprint %s not found: %w", id, models.ErrNotFound)
	}
	delete(r.byID, id)
	return nil
}

func (r *SprintRepository) GetByID(_ context.Context, id uuid.UUID) (*models.Sprint, error) {
	s, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	cp := *s
	return &cp, nil
}

func (r *SprintRepository) List(_ context.Context, projectID uuid.UUID) ([]*models.Sprint, error) {
	var result []*models.Sprint
	for _, s := range r.byID {
		if s.ProjectID == projectID {
			cp := *s
			result = append(result, &cp)
		}
	}
	return result, nil
}
