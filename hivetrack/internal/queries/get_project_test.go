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

func TestHandleGetProject_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))

	project := &models.Project{ID: uuid.New(), Slug: "backend", Name: "Backend", Archetype: models.ProjectArchetypeSoftware, CreatedBy: actor.ID}
	require.NoError(t, db.Projects().Insert(context.Background(), project))
	require.NoError(t, db.Projects().AddMember(context.Background(), &models.ProjectMember{
		ProjectID: project.ID, UserID: actor.ID, Role: models.ProjectRoleAdmin,
	}))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetProject(ctx, queries.GetProjectQuery{Slug: "backend"})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "backend", result.Slug)
	assert.Len(t, result.Members, 1)
}

func TestHandleGetProject_NotFound(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetProject(ctx, queries.GetProjectQuery{Slug: "nope"})
	require.NoError(t, err)
	assert.Nil(t, result)
}
