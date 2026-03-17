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

func TestHandleDeleteProject_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	actor.SetIsAdmin(true)
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	project := models.NewProject(actor.GetId(), "del-me", "Delete Me", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleDeleteProject(ctx, commands.DeleteProjectCommand{ID: project.GetId()})
	require.NoError(t, err)

	deleted, err := db.Projects().GetByID(context.Background(), project.GetId())
	require.NoError(t, err)
	assert.Nil(t, deleted)
}
