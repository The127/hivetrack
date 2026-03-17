package queries_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/queries"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

func TestHandleGetSprints(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	sprint := models.NewSprint(p.GetId(), "Sprint 1", nil, time.Now(), time.Now().Add(14*24*time.Hour), models.SprintStatusPlanning)
	db.Sprints().Insert(sprint)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetSprints(ctx, queries.GetSprintsQuery{ProjectID: p.GetId()})
	require.NoError(t, err)
	assert.Len(t, result.Sprints, 1)
	assert.Equal(t, "Sprint 1", result.Sprints[0].Name)
}
