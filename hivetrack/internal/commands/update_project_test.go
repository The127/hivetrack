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

func TestHandleUpdateProject_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com", IsAdmin: true}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))

	project := &models.Project{
		ID:        uuid.New(),
		Slug:      "my-project",
		Name:      "Old Name",
		Archetype: models.ProjectArchetypeSoftware,
		CreatedBy: actor.ID,
	}
	require.NoError(t, db.Projects().Insert(context.Background(), project))
	require.NoError(t, db.Projects().AddMember(context.Background(), &models.ProjectMember{
		ProjectID: project.ID, UserID: actor.ID, Role: models.ProjectRoleAdmin,
	}))

	newName := "New Name"
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	_, err := commands.HandleUpdateProject(ctx, commands.UpdateProjectCommand{
		ID:   project.ID,
		Name: &newName,
	})
	require.NoError(t, err)

	updated, err := db.Projects().GetByID(context.Background(), project.ID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
}

func TestHandleUpdateProject_Archive(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com", IsAdmin: true}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))

	project := &models.Project{
		ID:        uuid.New(),
		Slug:      "p1",
		Name:      "Project",
		Archetype: models.ProjectArchetypeSoftware,
		CreatedBy: actor.ID,
	}
	require.NoError(t, db.Projects().Insert(context.Background(), project))

	archived := true
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleUpdateProject(ctx, commands.UpdateProjectCommand{
		ID:       project.ID,
		Archived: &archived,
	})
	require.NoError(t, err)

	updated, err := db.Projects().GetByID(context.Background(), project.ID)
	require.NoError(t, err)
	assert.True(t, updated.Archived)
}

func TestHandleUpdateProject_NotFound(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleUpdateProject(ctx, commands.UpdateProjectCommand{
		ID: uuid.New(),
	})
	require.Error(t, err)
}
