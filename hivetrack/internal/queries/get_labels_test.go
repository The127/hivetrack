package queries_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/queries"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

func TestHandleGetLabels(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))

	p := &models.Project{ID: uuid.New(), Slug: "p", Name: "P", Archetype: models.ProjectArchetypeSoftware, CreatedBy: actor.ID}
	require.NoError(t, db.Projects().Insert(context.Background(), p))

	l := &models.Label{ID: uuid.New(), ProjectID: p.ID, Name: "bug", Color: "#ff0000"}
	require.NoError(t, db.Labels().Insert(context.Background(), l))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetLabels(ctx, queries.GetLabelsQuery{ProjectID: p.ID})
	require.NoError(t, err)
	assert.Len(t, result.Labels, 1)
	assert.Equal(t, "bug", result.Labels[0].Name)
}
