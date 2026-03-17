package inmemory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
)

type MilestoneRepository struct {
	byID map[uuid.UUID]*models.Milestone
}

func NewMilestoneRepository() *MilestoneRepository {
	return &MilestoneRepository{
		byID: make(map[uuid.UUID]*models.Milestone),
	}
}

func (r *MilestoneRepository) Insert(_ context.Context, milestone *models.Milestone) error {
	cp := *milestone
	r.byID[milestone.ID] = &cp
	return nil
}

func (r *MilestoneRepository) Update(_ context.Context, milestone *models.Milestone) error {
	if _, ok := r.byID[milestone.ID]; !ok {
		return fmt.Errorf("milestone %s not found: %w", milestone.ID, models.ErrNotFound)
	}
	cp := *milestone
	r.byID[milestone.ID] = &cp
	return nil
}

func (r *MilestoneRepository) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := r.byID[id]; !ok {
		return fmt.Errorf("milestone %s not found: %w", id, models.ErrNotFound)
	}
	delete(r.byID, id)
	return nil
}

func (r *MilestoneRepository) GetByID(_ context.Context, id uuid.UUID) (*models.Milestone, error) {
	m, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	cp := *m
	return &cp, nil
}

func (r *MilestoneRepository) List(_ context.Context, projectID uuid.UUID) ([]*models.Milestone, error) {
	var result []*models.Milestone
	for _, m := range r.byID {
		if m.ProjectID == projectID {
			cp := *m
			result = append(result, &cp)
		}
	}
	return result, nil
}
