package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/repositories"
)

// UserInfo is a lightweight user representation for embedding in query responses.
type UserInfo struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
}

// resolveUsers batch-fetches user details for a set of UUIDs and returns them as UserInfo slices.
// Unknown IDs are silently skipped.
func resolveUsers(ctx context.Context, db repositories.DbContext, ids []uuid.UUID) ([]UserInfo, error) {
	if len(ids) == 0 {
		return []UserInfo{}, nil
	}

	result := make([]UserInfo, 0, len(ids))
	for _, id := range ids {
		user, err := db.Users().GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if user == nil {
			continue
		}
		result = append(result, UserInfo{
			ID:          user.GetId(),
			DisplayName: user.GetDisplayName(),
			AvatarURL:   user.GetAvatarURL(),
		})
	}
	return result, nil
}
