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

func TestHandleGetMilestones(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	m := models.NewMilestone(p.GetId(), "v1.0", nil, nil)
	db.Milestones().Insert(m)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetMilestones(ctx, queries.GetMilestonesQuery{ProjectSlug: "p"})
	require.NoError(t, err)
	assert.Len(t, result.Milestones, 1)
	assert.Equal(t, "v1.0", result.Milestones[0].Title)
}

func TestHandleGetMilestones_WithProgress(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "progress-test", "Progress Test", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	m := models.NewMilestone(p.GetId(), "v2.0", nil, nil)
	db.Milestones().Insert(m)
	require.NoError(t, db.SaveChanges(context.Background()))

	mid := m.GetId()
	i1 := models.NewIssue(p.GetId(), 1, models.IssueTypeTask, "issue 1",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		nil, false, models.IssueVisibilityNormal,
		nil, nil, &mid, nil, nil)
	i2 := models.NewIssue(p.GetId(), 2, models.IssueTypeTask, "issue 2",
		models.IssueStatusDone, models.IssuePriorityNone, models.IssueEstimateNone,
		nil, false, models.IssueVisibilityNormal,
		nil, nil, &mid, nil, nil)
	db.Issues().Insert(i1)
	db.Issues().Insert(i2)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetMilestones(ctx, queries.GetMilestonesQuery{ProjectSlug: "progress-test"})
	require.NoError(t, err)
	require.Len(t, result.Milestones, 1)
	assert.Equal(t, 2, result.Milestones[0].IssueCount)
	assert.Equal(t, 1, result.Milestones[0].ClosedIssueCount)
}
