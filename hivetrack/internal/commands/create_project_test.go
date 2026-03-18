package commands_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

func TestHandleCreateProject_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	result, err := commands.HandleCreateProject(ctx, commands.CreateProjectCommand{
		Slug:      "backend",
		Name:      "Backend Platform",
		Archetype: models.ProjectArchetypeSoftware,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "backend", result.Slug)

	// Project should be persisted
	project, err := db.Projects().GetBySlug(context.Background(), "backend")
	require.NoError(t, err)
	require.NotNil(t, project)
	assert.Equal(t, "Backend Platform", project.GetName())
	assert.Equal(t, models.ProjectArchetypeSoftware, project.GetArchetype())
	assert.Equal(t, actor.GetId(), project.GetCreatedBy())

	// Creator should be added as project_admin
	member, err := db.Projects().GetMember(context.Background(), project.GetId(), actor.GetId())
	require.NoError(t, err)
	require.NotNil(t, member)
	assert.Equal(t, models.ProjectRoleAdmin, member.Role)
}

func TestHandleCreateProject_WithDescription(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	desc := "A project description"
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	result, err := commands.HandleCreateProject(ctx, commands.CreateProjectCommand{
		Slug:        "myproject",
		Name:        "My Project",
		Archetype:   models.ProjectArchetypeSupport,
		Description: &desc,
	})

	require.NoError(t, err)
	require.NotNil(t, result)

	project, err := db.Projects().GetByID(context.Background(), result.ID)
	require.NoError(t, err)
	require.NotNil(t, project.GetDescription())
	assert.Equal(t, desc, *project.GetDescription())
}
