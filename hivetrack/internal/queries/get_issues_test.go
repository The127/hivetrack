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

func seedIssue(db *inmemory.DbContext, projectID, reporterID uuid.UUID, number int, status models.IssueStatus, triaged bool) *models.Issue {
	issue := models.NewIssue(
		projectID, number, models.IssueTypeTask, "Issue "+string(status),
		status, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, triaged, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(issue)
	_ = db.SaveChanges(context.Background())
	return issue
}

func TestHandleGetIssues_ByProject(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	seedIssue(db, p.GetId(), actor.GetId(), 1, models.IssueStatusTodo, true)
	seedIssue(db, p.GetId(), actor.GetId(), 2, models.IssueStatusInProgress, true)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetIssues(ctx, queries.GetIssuesQuery{ProjectSlug: "p"})
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	assert.Len(t, result.Items, 2)
}

func TestHandleGetIssues_FilterByStatus(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	seedIssue(db, p.GetId(), actor.GetId(), 1, models.IssueStatusTodo, true)
	seedIssue(db, p.GetId(), actor.GetId(), 2, models.IssueStatusDone, true)

	status := models.IssueStatusTodo
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetIssues(ctx, queries.GetIssuesQuery{
		ProjectSlug: "p",
		Status:      &status,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
}
