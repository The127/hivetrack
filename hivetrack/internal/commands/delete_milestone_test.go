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

func TestHandleDeleteMilestone_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "Test User")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "delete-ms", "Delete MS", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	m := models.NewMilestone(p.GetId(), "v1.0", nil, nil)
	db.Milestones().Insert(m)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleDeleteMilestone(ctx, commands.DeleteMilestoneCommand{
		MilestoneID: m.GetId(),
	})

	require.NoError(t, err)

	deleted, err := db.Milestones().GetByID(context.Background(), m.GetId())
	require.NoError(t, err)
	assert.Nil(t, deleted)
}

func TestHandleDeleteMilestone_NotFound(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "Test User")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleDeleteMilestone(ctx, commands.DeleteMilestoneCommand{
		MilestoneID: [16]byte{0x01},
	})

	require.ErrorIs(t, err, models.ErrNotFound)
}
