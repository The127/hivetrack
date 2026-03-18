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

func TestHandleUpdateMilestone_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "Test User")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "update-ms", "Update MS", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	m := models.NewMilestone(p.GetId(), "v1.0", nil, nil)
	db.Milestones().Insert(m)
	require.NoError(t, db.SaveChanges(context.Background()))

	newTitle := "v1.1"
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleUpdateMilestone(ctx, commands.UpdateMilestoneCommand{
		MilestoneID: m.GetId(),
		Title:       &newTitle,
	})

	require.NoError(t, err)

	updated, err := db.Milestones().GetByID(context.Background(), m.GetId())
	require.NoError(t, err)
	assert.Equal(t, "v1.1", updated.GetTitle())
}

func TestHandleUpdateMilestone_Close(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "Test User")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "close-ms", "Close MS", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	m := models.NewMilestone(p.GetId(), "v1.0", nil, nil)
	db.Milestones().Insert(m)
	require.NoError(t, db.SaveChanges(context.Background()))

	closeIt := true
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleUpdateMilestone(ctx, commands.UpdateMilestoneCommand{
		MilestoneID: m.GetId(),
		Close:       &closeIt,
	})

	require.NoError(t, err)

	updated, err := db.Milestones().GetByID(context.Background(), m.GetId())
	require.NoError(t, err)
	assert.NotNil(t, updated.GetClosedAt())
}

func TestHandleUpdateMilestone_NotFound(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "Test User")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	title := "v2"
	_, err := commands.HandleUpdateMilestone(ctx, commands.UpdateMilestoneCommand{
		MilestoneID: [16]byte{0x01},
		Title:       &title,
	})

	require.ErrorIs(t, err, models.ErrNotFound)
}
