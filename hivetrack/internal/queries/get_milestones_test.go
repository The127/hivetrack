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

func TestHandleGetMilestones(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))

	p := &models.Project{ID: uuid.New(), Slug: "p", Name: "P", Archetype: models.ProjectArchetypeSoftware, CreatedBy: actor.ID}
	require.NoError(t, db.Projects().Insert(context.Background(), p))

	m := &models.Milestone{
		ID:        uuid.New(),
		ProjectID: p.ID,
		Title:     "v1.0",
		CreatedAt: time.Now(),
	}
	require.NoError(t, db.Milestones().Insert(context.Background(), m))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetMilestones(ctx, queries.GetMilestonesQuery{ProjectID: p.ID})
	require.NoError(t, err)
	assert.Len(t, result.Milestones, 1)
	assert.Equal(t, "v1.0", result.Milestones[0].Title)
}
