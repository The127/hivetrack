//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories/postgres"
)

func TestUserRepository_UpsertAndGetByID(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	sub := fmt.Sprintf("sub-%s", uuid.New())
	user := models.NewUser(sub, fmt.Sprintf("%s@test.com", sub), "Test User")

	err := db.Users().Upsert(ctx, user)
	require.NoError(t, err)

	got, err := db.Users().GetByID(ctx, user.GetId())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, sub, got.GetSub())
	assert.Equal(t, "Test User", got.GetDisplayName())
}

func TestUserRepository_GetBySub(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	sub := fmt.Sprintf("sub-%s", uuid.New())
	user := models.NewUser(sub, fmt.Sprintf("%s@test.com", sub), "Sub User")
	require.NoError(t, db.Users().Upsert(ctx, user))

	got, err := db.Users().GetBySub(ctx, sub)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, user.GetId(), got.GetId())
}

func TestUserRepository_GetByEmail(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	sub := fmt.Sprintf("sub-%s", uuid.New())
	email := fmt.Sprintf("%s@test.com", sub)
	user := models.NewUser(sub, email, "Email User")
	require.NoError(t, db.Users().Upsert(ctx, user))

	got, err := db.Users().GetByEmail(ctx, email)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, user.GetId(), got.GetId())
}

func TestUserRepository_Upsert_UpdatesExisting(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	sub := fmt.Sprintf("sub-%s", uuid.New())
	user := models.NewUser(sub, fmt.Sprintf("%s@test.com", sub), "Original Name")
	require.NoError(t, db.Users().Upsert(ctx, user))

	user.SetDisplayName("Updated Name")
	require.NoError(t, db.Users().Upsert(ctx, user))

	got, err := db.Users().GetBySub(ctx, sub)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Updated Name", got.GetDisplayName())
}

func TestUserRepository_List(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	sub := fmt.Sprintf("sub-%s", uuid.New())
	user := models.NewUser(sub, fmt.Sprintf("%s@test.com", sub), "List User")
	require.NoError(t, db.Users().Upsert(ctx, user))

	users, err := db.Users().List(ctx)
	require.NoError(t, err)

	var found bool
	for _, u := range users {
		if u.GetId() == user.GetId() {
			found = true
			break
		}
	}
	assert.True(t, found, "created user should appear in list")
}
