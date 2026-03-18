package queries_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/queries"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

func TestHandleGetSprintBurndown_HappyPath(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "sw", "SW", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	start := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)
	sprint := models.NewSprint(p.GetId(), 1, "Sprint 1", nil, start, end, models.SprintStatusCompleted)
	db.Sprints().Insert(sprint)
	require.NoError(t, db.SaveChanges(context.Background()))

	sprintID := sprint.GetId()
	reporterID := actor.GetId()

	// 3 task issues in sprint
	issue1 := models.NewIssue(p.GetId(), 1, models.IssueTypeTask, "Issue 1",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal, nil, &sprintID, nil, nil, nil)
	issue2 := models.NewIssue(p.GetId(), 2, models.IssueTypeTask, "Issue 2",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal, nil, &sprintID, nil, nil, nil)
	issue3 := models.NewIssue(p.GetId(), 3, models.IssueTypeTask, "Issue 3",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal, nil, &sprintID, nil, nil, nil)
	db.Issues().Insert(issue1)
	db.Issues().Insert(issue2)
	db.Issues().Insert(issue3)
	require.NoError(t, db.SaveChanges(context.Background()))

	// issue1 done on Mar 11 (day 2), issue2 done on Mar 13 (day 4)
	day2 := time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC)
	day4 := time.Date(2026, 3, 13, 0, 0, 0, 0, time.UTC)
	ctx := context.Background()
	require.NoError(t, db.IssueStatusLog().Insert(ctx, issue1.GetId(), "done", day2))
	require.NoError(t, db.IssueStatusLog().Insert(ctx, issue2.GetId(), "done", day4))

	queryCtx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetSprintBurndown(queryCtx, queries.GetSprintBurndownQuery{
		ProjectSlug: "sw",
		SprintID:    sprint.GetId(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 3, result.Total)

	pointsByDate := make(map[string]int)
	for _, pt := range result.Points {
		pointsByDate[pt.Date.Format("2006-01-02")] = pt.Remaining
	}

	// Mar 10: all 3 remaining (no terminal entries yet)
	assert.Equal(t, 3, pointsByDate["2026-03-10"])
	// Mar 11: issue1 completed → 2 remaining
	assert.Equal(t, 2, pointsByDate["2026-03-11"])
	// Mar 12: still 2 remaining
	assert.Equal(t, 2, pointsByDate["2026-03-12"])
	// Mar 13: issue2 completed → 1 remaining
	assert.Equal(t, 1, pointsByDate["2026-03-13"])
	// Mar 14–16: 1 remaining
	assert.Equal(t, 1, pointsByDate["2026-03-14"])
}

func TestHandleGetSprintBurndown_EmptyStatusLog(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "sw2", "SW2", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	start := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC)
	sprint := models.NewSprint(p.GetId(), 2, "Sprint 2", nil, start, end, models.SprintStatusCompleted)
	db.Sprints().Insert(sprint)
	require.NoError(t, db.SaveChanges(context.Background()))

	sprintID := sprint.GetId()
	reporterID := actor.GetId()

	issue1 := models.NewIssue(p.GetId(), 1, models.IssueTypeTask, "Issue 1",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal, nil, &sprintID, nil, nil, nil)
	issue2 := models.NewIssue(p.GetId(), 2, models.IssueTypeTask, "Issue 2",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal, nil, &sprintID, nil, nil, nil)
	db.Issues().Insert(issue1)
	db.Issues().Insert(issue2)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetSprintBurndown(ctx, queries.GetSprintBurndownQuery{
		ProjectSlug: "sw2",
		SprintID:    sprint.GetId(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Total)

	// No terminal entries → all days show 2 remaining
	for _, pt := range result.Points {
		assert.Equal(t, 2, pt.Remaining, "expected all remaining on %s", pt.Date.Format("2006-01-02"))
	}
}

func TestHandleGetSprintBurndown_SprintNotFound(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "sw3", "SW3", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := queries.HandleGetSprintBurndown(ctx, queries.GetSprintBurndownQuery{
		ProjectSlug: "sw3",
		SprintID:    [16]byte{1, 2, 3},
	})
	require.Error(t, err)
}
