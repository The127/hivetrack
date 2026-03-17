package commands_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

func setupSprintTest(t *testing.T) (*inmemory.DbContext, *models.Project, *models.User) {
	t.Helper()
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "sp", "Sprint Project", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))
	return db, project, actor
}

func makeIssue(projectID uuid.UUID, number int, status models.IssueStatus, sprintID *uuid.UUID, reporterID uuid.UUID) *models.Issue {
	rid := reporterID
	return models.NewIssue(
		projectID, number, models.IssueTypeTask, "Issue",
		status, models.IssuePriorityNone, models.IssueEstimateNone,
		&rid, true, models.IssueVisibilityNormal,
		nil, sprintID, nil, nil, nil,
	)
}

func TestHandleUpdateSprint_Complete_MovesToBacklog(t *testing.T) {
	db, project, actor := setupSprintTest(t)

	sprint := models.NewSprint(project.GetId(), "Sprint 1", nil, time.Now(), time.Now().Add(14*24*time.Hour), models.SprintStatusActive)
	db.Sprints().Insert(sprint)
	require.NoError(t, db.SaveChanges(context.Background()))

	sid1 := sprint.GetId()
	sid2 := sprint.GetId()
	openIssue := makeIssue(project.GetId(), 1, models.IssueStatusInProgress, &sid1, actor.GetId())
	doneIssue := makeIssue(project.GetId(), 2, models.IssueStatusDone, &sid2, actor.GetId())
	db.Issues().Insert(openIssue)
	db.Issues().Insert(doneIssue)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	status := models.SprintStatusCompleted
	_, err := commands.HandleUpdateSprint(ctx, commands.UpdateSprintCommand{
		SprintID: sprint.GetId(),
		Status:   &status,
	})
	require.NoError(t, err)

	updated, err := db.Issues().GetByID(context.Background(), openIssue.GetId())
	require.NoError(t, err)
	assert.Nil(t, updated.GetSprintID())
	assert.Equal(t, 1, updated.GetSprintCarryCount())

	done, err := db.Issues().GetByID(context.Background(), doneIssue.GetId())
	require.NoError(t, err)
	require.NotNil(t, done.GetSprintID())
	assert.Equal(t, sprint.GetId(), *done.GetSprintID())
}

func TestHandleUpdateSprint_Complete_MovesToAnotherSprint(t *testing.T) {
	db, project, actor := setupSprintTest(t)

	sprint := models.NewSprint(project.GetId(), "Sprint 1", nil, time.Now(), time.Now().Add(14*24*time.Hour), models.SprintStatusActive)
	nextSprint := models.NewSprint(project.GetId(), "Sprint 2", nil, time.Now().Add(14*24*time.Hour), time.Now().Add(28*24*time.Hour), models.SprintStatusPlanning)
	db.Sprints().Insert(sprint)
	db.Sprints().Insert(nextSprint)
	require.NoError(t, db.SaveChanges(context.Background()))

	sid1 := sprint.GetId()
	sid2 := sprint.GetId()
	openIssue := makeIssue(project.GetId(), 1, models.IssueStatusInProgress, &sid1, actor.GetId())
	doneIssue := makeIssue(project.GetId(), 2, models.IssueStatusDone, &sid2, actor.GetId())
	db.Issues().Insert(openIssue)
	db.Issues().Insert(doneIssue)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	status := models.SprintStatusCompleted
	nextID := nextSprint.GetId()
	_, err := commands.HandleUpdateSprint(ctx, commands.UpdateSprintCommand{
		SprintID:                 sprint.GetId(),
		Status:                   &status,
		MoveOpenIssuesToSprintID: &nextID,
	})
	require.NoError(t, err)

	updated, err := db.Issues().GetByID(context.Background(), openIssue.GetId())
	require.NoError(t, err)
	require.NotNil(t, updated.GetSprintID())
	assert.Equal(t, nextSprint.GetId(), *updated.GetSprintID())
	assert.Equal(t, 1, updated.GetSprintCarryCount())

	done, err := db.Issues().GetByID(context.Background(), doneIssue.GetId())
	require.NoError(t, err)
	require.NotNil(t, done.GetSprintID())
	assert.Equal(t, sprint.GetId(), *done.GetSprintID())
}

func TestHandleUpdateSprint_Complete_InvalidTargetSprint(t *testing.T) {
	db, project, actor := setupSprintTest(t)

	sprint := models.NewSprint(project.GetId(), "Sprint 1", nil, time.Now(), time.Now().Add(14*24*time.Hour), models.SprintStatusActive)
	db.Sprints().Insert(sprint)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	status := models.SprintStatusCompleted
	bogusID := uuid.New()
	_, err := commands.HandleUpdateSprint(ctx, commands.UpdateSprintCommand{
		SprintID:                 sprint.GetId(),
		Status:                   &status,
		MoveOpenIssuesToSprintID: &bogusID,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, models.ErrNotFound)
}

func TestHandleUpdateSprint_Complete_CannotMoveToSameSprint(t *testing.T) {
	db, project, actor := setupSprintTest(t)

	sprint := models.NewSprint(project.GetId(), "Sprint 1", nil, time.Now(), time.Now().Add(14*24*time.Hour), models.SprintStatusActive)
	db.Sprints().Insert(sprint)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	status := models.SprintStatusCompleted
	sameID := sprint.GetId()
	_, err := commands.HandleUpdateSprint(ctx, commands.UpdateSprintCommand{
		SprintID:                 sprint.GetId(),
		Status:                   &status,
		MoveOpenIssuesToSprintID: &sameID,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, models.ErrBadRequest)
}
