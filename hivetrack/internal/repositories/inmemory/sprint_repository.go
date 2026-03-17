package inmemory

import (
	"context"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/change"
	"github.com/the127/hivetrack/internal/models"
)

type SprintRepository struct {
	tracker *change.Tracker
	byID    map[uuid.UUID]*models.Sprint
}

func NewSprintRepository(tracker *change.Tracker) *SprintRepository {
	return &SprintRepository{
		tracker: tracker,
		byID:    make(map[uuid.UUID]*models.Sprint),
	}
}

func (r *SprintRepository) Insert(sprint *models.Sprint) {
	r.tracker.Add(change.NewEntry(0, sprint, change.Added))
}

func (r *SprintRepository) Update(sprint *models.Sprint) {
	r.tracker.Add(change.NewEntry(0, sprint, change.Updated))
}

func (r *SprintRepository) Delete(sprint *models.Sprint) {
	r.tracker.Add(change.NewEntry(0, sprint, change.Deleted))
}

func (r *SprintRepository) GetByID(_ context.Context, id uuid.UUID) (*models.Sprint, error) {
	s, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (r *SprintRepository) List(_ context.Context, projectID uuid.UUID) ([]*models.Sprint, error) {
	var result []*models.Sprint
	for _, s := range r.byID {
		if s.GetProjectID() == projectID {
			result = append(result, s)
		}
	}
	return result, nil
}
