package inmemory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
)

type UserRepository struct {
	byID    map[uuid.UUID]*models.User
	bySub   map[string]*models.User
	byEmail map[string]*models.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		byID:    make(map[uuid.UUID]*models.User),
		bySub:   make(map[string]*models.User),
		byEmail: make(map[string]*models.User),
	}
}

func (r *UserRepository) GetByID(_ context.Context, id uuid.UUID) (*models.User, error) {
	u, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (r *UserRepository) GetBySub(_ context.Context, sub string) (*models.User, error) {
	u, ok := r.bySub[sub]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (r *UserRepository) GetByEmail(_ context.Context, email string) (*models.User, error) {
	u, ok := r.byEmail[email]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (r *UserRepository) Upsert(_ context.Context, user *models.User) error {
	if user.GetId() == uuid.Nil {
		return fmt.Errorf("user ID must be set")
	}
	// Remove old email/sub indexes if updating
	if existing, ok := r.byID[user.GetId()]; ok {
		delete(r.bySub, existing.GetSub())
		delete(r.byEmail, existing.GetEmail())
	}
	r.byID[user.GetId()] = user
	r.bySub[user.GetSub()] = user
	r.byEmail[user.GetEmail()] = user
	return nil
}

func (r *UserRepository) List(_ context.Context) ([]*models.User, error) {
	users := make([]*models.User, 0, len(r.byID))
	for _, u := range r.byID {
		users = append(users, u)
	}
	return users, nil
}
