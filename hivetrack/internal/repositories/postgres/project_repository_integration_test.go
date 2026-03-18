//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
	"github.com/the127/hivetrack/internal/repositories/postgres"
)

func newTestProject(t *testing.T, createdBy uuid.UUID) *models.Project {
	t.Helper()
	slug := fmt.Sprintf("inttest-%s", uuid.New().String()[:8])
	return models.NewProject(createdBy, slug, "Integration Test Project", models.ProjectArchetypeSoftware)
}

func newTestUser(t *testing.T, ctx context.Context, db *postgres.DbContext) *models.User {
	t.Helper()
	sub := fmt.Sprintf("sub-%s", uuid.New())
	user := models.NewUser(sub, fmt.Sprintf("%s@test.com", sub), "Test User")
	require.NoError(t, db.Users().Upsert(ctx, user))
	return user
}

func TestProjectRepository_InsertAndGetByID(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user := newTestUser(t, ctx, db)
	project := newTestProject(t, user.GetId())

	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(ctx))

	got, err := db.Projects().GetByID(ctx, project.GetId())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, project.GetSlug(), got.GetSlug())
	assert.Equal(t, project.GetName(), got.GetName())
	assert.Equal(t, models.ProjectArchetypeSoftware, got.GetArchetype())
}

func TestProjectRepository_GetBySlug(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user := newTestUser(t, ctx, db)
	project := newTestProject(t, user.GetId())

	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(ctx))

	got, err := db.Projects().GetBySlug(ctx, project.GetSlug())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, project.GetId(), got.GetId())
}

func TestProjectRepository_GetBySlug_NotFound(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	got, err := db.Projects().GetBySlug(ctx, "no-such-slug-"+uuid.New().String())
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestProjectRepository_Update(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user := newTestUser(t, ctx, db)
	project := newTestProject(t, user.GetId())
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(ctx))

	project.SetName("Renamed Project")
	db.Projects().Update(project)
	require.NoError(t, db.SaveChanges(ctx))

	got, err := db.Projects().GetByID(ctx, project.GetId())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Renamed Project", got.GetName())
}

func TestProjectRepository_Delete(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user := newTestUser(t, ctx, db)
	project := newTestProject(t, user.GetId())
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(ctx))

	db.Projects().Delete(project)
	require.NoError(t, db.SaveChanges(ctx))

	got, err := db.Projects().GetByID(ctx, project.GetId())
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestProjectRepository_List_FilterByMember(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user := newTestUser(t, ctx, db)
	project := newTestProject(t, user.GetId())
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(ctx))

	err := db.Projects().AddMember(ctx, &models.ProjectMember{
		ProjectID: project.GetId(),
		UserID:    user.GetId(),
		Role:      models.ProjectRoleAdmin,
	})
	require.NoError(t, err)

	projects, err := db.Projects().List(ctx, repositories.NewProjectFilter().ForMember(user.GetId()))
	require.NoError(t, err)

	var found bool
	for _, p := range projects {
		if p.GetId() == project.GetId() {
			found = true
			break
		}
	}
	assert.True(t, found, "project should appear in member's project list")
}

func TestProjectRepository_Members(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user := newTestUser(t, ctx, db)
	project := newTestProject(t, user.GetId())
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(ctx))

	// AddMember
	member := &models.ProjectMember{
		ProjectID: project.GetId(),
		UserID:    user.GetId(),
		Role:      models.ProjectRoleMember,
	}
	require.NoError(t, db.Projects().AddMember(ctx, member))

	// GetMember
	got, err := db.Projects().GetMember(ctx, project.GetId(), user.GetId())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, models.ProjectRoleMember, got.Role)

	// UpdateMember
	member.Role = models.ProjectRoleAdmin
	require.NoError(t, db.Projects().UpdateMember(ctx, member))

	got, err = db.Projects().GetMember(ctx, project.GetId(), user.GetId())
	require.NoError(t, err)
	assert.Equal(t, models.ProjectRoleAdmin, got.Role)

	// ListMembers
	members, err := db.Projects().ListMembers(ctx, project.GetId())
	require.NoError(t, err)
	assert.Len(t, members, 1)

	// RemoveMember
	require.NoError(t, db.Projects().RemoveMember(ctx, project.GetId(), user.GetId()))

	got, err = db.Projects().GetMember(ctx, project.GetId(), user.GetId())
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestProjectRepository_NextIssueNumber(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user := newTestUser(t, ctx, db)
	project := newTestProject(t, user.GetId())
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(ctx))

	n1, err := db.Projects().NextIssueNumber(ctx, project.GetId())
	require.NoError(t, err)

	n2, err := db.Projects().NextIssueNumber(ctx, project.GetId())
	require.NoError(t, err)

	assert.Greater(t, n2, n1, "issue numbers must increment")
}
