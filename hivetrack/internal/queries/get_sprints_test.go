package queries_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/queries"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

func TestHandleGetSprints(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))

	p := &models.Project{ID: uuid.New(), Slug: "p", Name: "P", Archetype: models.ProjectArchetypeSoftware, CreatedBy: actor.ID}
	require.NoError(t, db.Projects().Insert(context.Background(), p))

	sprint := &models.Sprint{
		ID:        uuid.New(),
		ProjectID: p.ID,
		Name:      "Sprint 1",
		StartDate: time.Now(),
		EndDate:   time.Now().Add(14 * 24 * time.Hour),
		Status:    models.SprintStatusPlanning,
		CreatedAt: time.Now(),
	}
	require.NoError(t, db.Sprints().Insert(context.Background(), sprint))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetSprints(ctx, queries.GetSprintsQuery{ProjectID: p.ID})
	require.NoError(t, err)
	assert.Len(t, result.Sprints, 1)
	assert.Equal(t, "Sprint 1", result.Sprints[0].Name)
}
