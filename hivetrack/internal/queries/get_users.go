package queries

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/repositories"
)

type GetUsersQuery struct{}

type UserSummary struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	IsAdmin     bool      `json:"is_admin"`
}

type GetUsersResult struct {
	Items []UserSummary `json:"items"`
}

func HandleGetUsers(ctx context.Context, _ GetUsersQuery) (*GetUsersResult, error) {
	db := repositories.GetDbContext(ctx)

	users, err := db.Users().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}

	summaries := make([]UserSummary, 0, len(users))
	for _, u := range users {
		summaries = append(summaries, UserSummary{
			ID:          u.GetId(),
			Email:       u.GetEmail(),
			DisplayName: u.GetDisplayName(),
			IsAdmin:     u.GetIsAdmin(),
		})
	}

	return &GetUsersResult{Items: summaries}, nil
}
