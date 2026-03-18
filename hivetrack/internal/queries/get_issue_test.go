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

func TestHandleGetIssue_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "myproj", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	issue := seedIssue(db, p.GetId(), actor.GetId(), 1, models.IssueStatusTodo, true)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetIssue(ctx, queries.GetIssueQuery{
		ProjectSlug: "myproj",
		Number:      issue.GetNumber(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, issue.GetId(), result.ID)
	assert.Equal(t, issue.GetTitle(), result.Title)
}

func TestHandleGetIssue_EpicIncludesChildProgress(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "myproj", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	epic := models.NewIssue(
		p.GetId(), 1, models.IssueTypeEpic, "Epic",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		nil, true, models.IssueVisibilityNormal, nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(epic)
	require.NoError(t, db.SaveChanges(context.Background()))

	epicID := epic.GetId()
	child1 := models.NewIssue(
		p.GetId(), 2, models.IssueTypeTask, "Task 1",
		models.IssueStatusInProgress, models.IssuePriorityNone, models.IssueEstimateNone,
		nil, true, models.IssueVisibilityNormal, nil, nil, nil, nil, nil,
	)
	child1.SetParentID(&epicID)
	db.Issues().Insert(child1)

	child2 := models.NewIssue(
		p.GetId(), 3, models.IssueTypeTask, "Task 2",
		models.IssueStatusDone, models.IssuePriorityNone, models.IssueEstimateNone,
		nil, true, models.IssueVisibilityNormal, nil, nil, nil, nil, nil,
	)
	child2.SetParentID(&epicID)
	db.Issues().Insert(child2)

	child3 := models.NewIssue(
		p.GetId(), 4, models.IssueTypeTask, "Task 3",
		models.IssueStatusCancelled, models.IssuePriorityNone, models.IssueEstimateNone,
		nil, true, models.IssueVisibilityNormal, nil, nil, nil, nil, nil,
	)
	child3.SetParentID(&epicID)
	db.Issues().Insert(child3)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetIssue(ctx, queries.GetIssueQuery{
		ProjectSlug: "myproj",
		Number:      1,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 3, result.ChildCount)
	assert.Equal(t, 2, result.ChildDoneCount) // done + cancelled
}

func TestHandleGetIssue_NotFound(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "myproj", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetIssue(ctx, queries.GetIssueQuery{
		ProjectSlug: "myproj",
		Number:      999,
	})
	require.NoError(t, err)
	assert.Nil(t, result)
}
