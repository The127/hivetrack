package inmemory

import (
	"context"
	"sort"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/change"
	"github.com/the127/hivetrack/internal/models"
)

type CommentRepository struct {
	tracker *change.Tracker
	byID    map[uuid.UUID]*models.Comment
}

func NewCommentRepository(tracker *change.Tracker) *CommentRepository {
	return &CommentRepository{
		tracker: tracker,
		byID:    make(map[uuid.UUID]*models.Comment),
	}
}

func (r *CommentRepository) Insert(comment *models.Comment) {
	r.tracker.Add(change.NewEntry(0, comment, change.Added))
}

func (r *CommentRepository) Update(comment *models.Comment) {
	r.tracker.Add(change.NewEntry(0, comment, change.Updated))
}

func (r *CommentRepository) Delete(comment *models.Comment) {
	r.tracker.Add(change.NewEntry(0, comment, change.Deleted))
}

func (r *CommentRepository) GetByID(_ context.Context, id uuid.UUID) (*models.Comment, error) {
	c, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	return c, nil
}

func (r *CommentRepository) List(_ context.Context, issueID uuid.UUID, limit, offset int) ([]*models.Comment, int, error) {
	var all []*models.Comment
	for _, c := range r.byID {
		if c.GetIssueID() == issueID {
			all = append(all, c)
		}
	}

	// Sort by created_at ascending
	sort.Slice(all, func(i, j int) bool {
		return all[i].GetCreatedAt().Before(all[j].GetCreatedAt())
	})

	total := len(all)

	if offset > len(all) {
		return nil, total, nil
	}
	all = all[offset:]
	if limit > 0 && limit < len(all) {
		all = all[:limit]
	}

	return all, total, nil
}
