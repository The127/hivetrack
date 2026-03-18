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

func TestHandleCreateMilestone_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "Test User")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "my-project", "My Project", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := commands.HandleCreateMilestone(ctx, commands.CreateMilestoneCommand{
		ProjectSlug: "my-project",
		Title:       "v1.0",
	})

	require.NoError(t, err)
	assert.NotEqual(t, result.ID, [16]byte{})

	m, err := db.Milestones().GetByID(context.Background(), result.ID)
	require.NoError(t, err)
	require.NotNil(t, m)
	assert.Equal(t, "v1.0", m.GetTitle())
	assert.Equal(t, p.GetId(), m.GetProjectID())
}

func TestHandleCreateMilestone_ProjectNotFound(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "Test User")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleCreateMilestone(ctx, commands.CreateMilestoneCommand{
		ProjectSlug: "nonexistent",
		Title:       "v1.0",
	})

	require.ErrorIs(t, err, models.ErrNotFound)
}
