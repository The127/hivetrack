package inmemory

import (
	"context"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/change"
	"github.com/the127/hivetrack/internal/models"
)

type LabelRepository struct {
	tracker *change.Tracker
	byID    map[uuid.UUID]*models.Label
}

func NewLabelRepository(tracker *change.Tracker) *LabelRepository {
	return &LabelRepository{
		tracker: tracker,
		byID:    make(map[uuid.UUID]*models.Label),
	}
}

func (r *LabelRepository) Insert(label *models.Label) {
	r.tracker.Add(change.NewEntry(0, label, change.Added))
}

func (r *LabelRepository) Update(label *models.Label) {
	r.tracker.Add(change.NewEntry(0, label, change.Updated))
}

func (r *LabelRepository) Delete(label *models.Label) {
	r.tracker.Add(change.NewEntry(0, label, change.Deleted))
}

func (r *LabelRepository) GetByID(_ context.Context, id uuid.UUID) (*models.Label, error) {
	l, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	return l, nil
}

func (r *LabelRepository) List(_ context.Context, projectID uuid.UUID) ([]*models.Label, error) {
	var result []*models.Label
	for _, l := range r.byID {
		if l.GetProjectID() == projectID {
			result = append(result, l)
		}
	}
	return result, nil
}
