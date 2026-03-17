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

func TestHandleGetProjects_MemberSeesOwnProjects(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	other := models.User{ID: uuid.New(), Sub: "sub2", Email: "other@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))
	require.NoError(t, db.Users().Upsert(context.Background(), &other))

	p1 := &models.Project{ID: uuid.New(), Slug: "p1", Name: "P1", Archetype: models.ProjectArchetypeSoftware, CreatedBy: actor.ID}
	p2 := &models.Project{ID: uuid.New(), Slug: "p2", Name: "P2", Archetype: models.ProjectArchetypeSoftware, CreatedBy: other.ID}
	require.NoError(t, db.Projects().Insert(context.Background(), p1))
	require.NoError(t, db.Projects().Insert(context.Background(), p2))

	// actor is member of p1 only
	require.NoError(t, db.Projects().AddMember(context.Background(), &models.ProjectMember{
		ProjectID: p1.ID, UserID: actor.ID, Role: models.ProjectRoleAdmin,
	}))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetProjects(ctx, queries.GetProjectsQuery{})
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "p1", result.Items[0].Slug)
}

func TestHandleGetProjects_AdminSeesAll(t *testing.T) {
	db := inmemory.NewDbContext()
	admin := models.User{ID: uuid.New(), Sub: "admin", Email: "admin@example.com", IsAdmin: true}
	require.NoError(t, db.Users().Upsert(context.Background(), &admin))

	p1 := &models.Project{ID: uuid.New(), Slug: "p1", Name: "P1", Archetype: models.ProjectArchetypeSoftware, CreatedBy: admin.ID}
	p2 := &models.Project{ID: uuid.New(), Slug: "p2", Name: "P2", Archetype: models.ProjectArchetypeSoftware, CreatedBy: admin.ID}
	require.NoError(t, db.Projects().Insert(context.Background(), p1))
	require.NoError(t, db.Projects().Insert(context.Background(), p2))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), admin)
	result, err := queries.HandleGetProjects(ctx, queries.GetProjectsQuery{})
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
}
