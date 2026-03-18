package queries_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/queries"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

func TestHandleGetProject_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	project := models.NewProject(actor.GetId(), "backend", "Backend", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))
	require.NoError(t, db.Projects().AddMember(context.Background(), &models.ProjectMember{
		ProjectID: project.GetId(), UserID: actor.GetId(), Role: models.ProjectRoleAdmin,
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
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetProject(ctx, queries.GetProjectQuery{Slug: "nope"})
	require.NoError(t, err)
	assert.Nil(t, result)
}
