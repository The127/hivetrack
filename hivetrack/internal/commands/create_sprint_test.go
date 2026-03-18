package commands_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

func TestHandleCreateSprint_AssignsSequentialNumbers(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "myproject", "My Project", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	now := time.Now()
	end := now.Add(14 * 24 * time.Hour)

	r1, err := commands.HandleCreateSprint(ctx, commands.CreateSprintCommand{
		ProjectSlug: project.GetSlug(),
		Name:        "Sprint 1",
		StartDate:   &now,
		EndDate:     &end,
	})
	require.NoError(t, err)

	r2, err := commands.HandleCreateSprint(ctx, commands.CreateSprintCommand{
		ProjectSlug: project.GetSlug(),
		Name:        "Sprint 2",
		StartDate:   &now,
		EndDate:     &end,
	})
	require.NoError(t, err)

	s1, err := db.Sprints().GetByID(context.Background(), r1.ID)
	require.NoError(t, err)
	s2, err := db.Sprints().GetByID(context.Background(), r2.ID)
	require.NoError(t, err)

	assert.Equal(t, 1, s1.GetNumber())
	assert.Equal(t, 2, s2.GetNumber())
}
