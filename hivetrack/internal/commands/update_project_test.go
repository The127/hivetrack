package commands_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

func TestHandleUpdateProject_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	actor.SetIsAdmin(true)
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	project := models.NewProject(actor.GetId(), "my-project", "Old Name", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))
	require.NoError(t, db.Projects().AddMember(context.Background(), &models.ProjectMember{
		ProjectID: project.GetId(), UserID: actor.GetId(), Role: models.ProjectRoleAdmin,
	}))

	newName := "New Name"
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	_, err := commands.HandleUpdateProject(ctx, commands.UpdateProjectCommand{
		ID:   project.GetId(),
		Name: &newName,
	})
	require.NoError(t, err)

	updated, err := db.Projects().GetByID(context.Background(), project.GetId())
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.GetName())
}

func TestHandleUpdateProject_Archive(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	actor.SetIsAdmin(true)
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	project := models.NewProject(actor.GetId(), "p1", "Project", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	archived := true
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleUpdateProject(ctx, commands.UpdateProjectCommand{
		ID:       project.GetId(),
		Archived: &archived,
	})
	require.NoError(t, err)

	updated, err := db.Projects().GetByID(context.Background(), project.GetId())
	require.NoError(t, err)
	assert.True(t, updated.GetArchived())
}

func TestHandleUpdateProject_NotFound(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	// We need a uuid.New() here but uuid isn't imported — use a known non-existent approach
	newProject := models.NewProject(actor.GetId(), "nonexistent", "N", models.ProjectArchetypeSoftware)
	_, err := commands.HandleUpdateProject(ctx, commands.UpdateProjectCommand{
		ID: newProject.GetId(),
	})
	require.Error(t, err)
}
