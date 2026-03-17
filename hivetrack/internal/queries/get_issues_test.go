package queries_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/queries"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

func seedIssue(db *inmemory.DbContext, projectID, reporterID uuid.UUID, number int, status models.IssueStatus, triaged bool) *models.Issue {
	issue := &models.Issue{
		ID:         uuid.New(),
		ProjectID:  projectID,
		Number:     number,
		Type:       models.IssueTypeTask,
		Title:      "Issue " + string(status),
		Status:     status,
		Priority:   models.IssuePriorityNone,
		Estimate:   models.IssueEstimateNone,
		ReporterID: &reporterID,
		Triaged:    triaged,
		Visibility: models.IssueVisibilityNormal,
		Checklist:  []models.ChecklistItem{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_ = db.Issues().Insert(context.Background(), issue)
	return issue
}

func TestHandleGetIssues_ByProject(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))

	p := &models.Project{ID: uuid.New(), Slug: "p", Name: "P", Archetype: models.ProjectArchetypeSoftware, CreatedBy: actor.ID}
	require.NoError(t, db.Projects().Insert(context.Background(), p))

	seedIssue(db, p.ID, actor.ID, 1, models.IssueStatusTodo, true)
	seedIssue(db, p.ID, actor.ID, 2, models.IssueStatusInProgress, true)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetIssues(ctx, queries.GetIssuesQuery{ProjectSlug: "p"})
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	assert.Len(t, result.Items, 2)
}

func TestHandleGetIssues_FilterByStatus(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))

	p := &models.Project{ID: uuid.New(), Slug: "p", Name: "P", Archetype: models.ProjectArchetypeSoftware, CreatedBy: actor.ID}
	require.NoError(t, db.Projects().Insert(context.Background(), p))

	seedIssue(db, p.ID, actor.ID, 1, models.IssueStatusTodo, true)
	seedIssue(db, p.ID, actor.ID, 2, models.IssueStatusDone, true)

	status := models.IssueStatusTodo
	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetIssues(ctx, queries.GetIssuesQuery{
		ProjectSlug: "p",
		Status:      &status,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
}
