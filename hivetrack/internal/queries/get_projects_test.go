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

func TestHandleGetProjects_MemberSeesOwnProjects(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	other := models.NewUser("sub2", "other@example.com", "other@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	require.NoError(t, db.Users().Upsert(context.Background(), other))

	p1 := models.NewProject(actor.GetId(), "p1", "P1", models.ProjectArchetypeSoftware)
	p2 := models.NewProject(other.GetId(), "p2", "P2", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p1)
	db.Projects().Insert(p2)
	require.NoError(t, db.SaveChanges(context.Background()))

	// actor is member of p1 only
	require.NoError(t, db.Projects().AddMember(context.Background(), &models.ProjectMember{
		ProjectID: p1.GetId(), UserID: actor.GetId(), Role: models.ProjectRoleAdmin,
	}))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetProjects(ctx, queries.GetProjectsQuery{})
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, "p1", result.Items[0].Slug)
}

func TestHandleGetProjects_AdminSeesAll(t *testing.T) {
	db := inmemory.NewDbContext()
	admin := models.NewUser("admin", "admin@example.com", "admin@example.com")
	admin.SetIsAdmin(true)
	require.NoError(t, db.Users().Upsert(context.Background(), admin))

	p1 := models.NewProject(admin.GetId(), "p1", "P1", models.ProjectArchetypeSoftware)
	p2 := models.NewProject(admin.GetId(), "p2", "P2", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p1)
	db.Projects().Insert(p2)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), admin)
	result, err := queries.HandleGetProjects(ctx, queries.GetProjectsQuery{})
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
}
