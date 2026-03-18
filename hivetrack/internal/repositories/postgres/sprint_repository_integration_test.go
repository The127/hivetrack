//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories/postgres"
)

func TestSprintRepository_InsertAndGetByID(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user := newTestUser(t, ctx, db)
	project := newTestProject(t, user.GetId())
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(ctx))

	goal := "Ship the feature"
	sprint := models.NewSprint(
		project.GetId(), "Sprint 1", &goal,
		time.Now().UTC().Truncate(24*time.Hour),
		time.Now().UTC().Truncate(24*time.Hour).Add(14*24*time.Hour),
		models.SprintStatusPlanning,
	)
	db.Sprints().Insert(sprint)
	require.NoError(t, db.SaveChanges(ctx))

	got, err := db.Sprints().GetByID(ctx, sprint.GetId())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Sprint 1", got.GetName())
	assert.Equal(t, project.GetId(), got.GetProjectID())
	assert.Equal(t, models.SprintStatusPlanning, got.GetStatus())
	require.NotNil(t, got.GetGoal())
	assert.Equal(t, "Ship the feature", *got.GetGoal())
}

func TestSprintRepository_List(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user := newTestUser(t, ctx, db)
	project := newTestProject(t, user.GetId())
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(ctx))

	s1 := models.NewSprint(project.GetId(), "Sprint A", nil,
		time.Now().UTC(), time.Now().UTC().Add(7*24*time.Hour), models.SprintStatusPlanning)
	s2 := models.NewSprint(project.GetId(), "Sprint B", nil,
		time.Now().UTC(), time.Now().UTC().Add(14*24*time.Hour), models.SprintStatusActive)
	db.Sprints().Insert(s1)
	db.Sprints().Insert(s2)
	require.NoError(t, db.SaveChanges(ctx))

	sprints, err := db.Sprints().List(ctx, project.GetId())
	require.NoError(t, err)
	assert.Len(t, sprints, 2)
}

func TestSprintRepository_Update(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user := newTestUser(t, ctx, db)
	project := newTestProject(t, user.GetId())
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(ctx))

	sprint := models.NewSprint(project.GetId(), "Sprint 1", nil,
		time.Now().UTC(), time.Now().UTC().Add(14*24*time.Hour), models.SprintStatusPlanning)
	db.Sprints().Insert(sprint)
	require.NoError(t, db.SaveChanges(ctx))

	sprint.SetStatus(models.SprintStatusActive)
	db.Sprints().Update(sprint)
	require.NoError(t, db.SaveChanges(ctx))

	got, err := db.Sprints().GetByID(ctx, sprint.GetId())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, models.SprintStatusActive, got.GetStatus())
}
