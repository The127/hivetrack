package commands_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

func TestHandleDeleteProject_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com", IsAdmin: true}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))

	project := &models.Project{
		ID:        uuid.New(),
		Slug:      "del-me",
		Name:      "Delete Me",
		Archetype: models.ProjectArchetypeSoftware,
		CreatedBy: actor.ID,
	}
	require.NoError(t, db.Projects().Insert(context.Background(), project))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleDeleteProject(ctx, commands.DeleteProjectCommand{ID: project.ID})
	require.NoError(t, err)

	deleted, err := db.Projects().GetByID(context.Background(), project.ID)
	require.NoError(t, err)
	assert.Nil(t, deleted)
}
