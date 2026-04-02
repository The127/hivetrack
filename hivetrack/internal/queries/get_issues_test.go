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

func TestHandleGetIssues_FilterByType(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	seedIssue(db, p.GetId(), actor.GetId(), 1, models.IssueStatusTodo, true)
	// Seed an epic
	actorID := actor.GetId()
	epic := models.NewIssue(
		p.GetId(), 2, models.IssueTypeEpic, "Epic issue",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&actorID, true, models.IssueVisibilityNormal, nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(epic)
	_ = db.SaveChanges(context.Background())

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	epicType := models.IssueTypeEpic
	result, err := queries.HandleGetIssues(ctx, queries.GetIssuesQuery{
		ProjectSlug: "p",
		Type:        &epicType,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, models.IssueTypeEpic, result.Items[0].Type)
}

func TestHandleGetIssues_FilterByParentID(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
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
	child := models.NewIssue(
		p.GetId(), 2, models.IssueTypeTask, "Child",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		nil, true, models.IssueVisibilityNormal, nil, nil, nil, nil, nil,
	)
	child.SetParentID(&epicID)
	db.Issues().Insert(child)

	orphan := models.NewIssue(
		p.GetId(), 3, models.IssueTypeTask, "Orphan",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		nil, true, models.IssueVisibilityNormal, nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(orphan)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetIssues(ctx, queries.GetIssuesQuery{
		ProjectSlug: "p",
		ParentID:    &epicID,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, "Child", result.Items[0].Title)
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

func TestHandleGetIssues_FilterByLabelID(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	labelID := uuid.New()
	otherLabelID := uuid.New()

	labeled := seedIssue(db, p.GetId(), actor.GetId(), 1, models.IssueStatusTodo, true)
	labeled.SetLabels([]uuid.UUID{labelID})
	db.Issues().Update(labeled)

	unlabeled := seedIssue(db, p.GetId(), actor.GetId(), 2, models.IssueStatusTodo, true)
	unlabeled.SetLabels([]uuid.UUID{otherLabelID})
	db.Issues().Update(unlabeled)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	// Filter by label
	result, err := queries.HandleGetIssues(ctx, queries.GetIssuesQuery{
		ProjectSlug: "p",
		LabelID:     &labelID,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 1, result.Items[0].Number)
}

func TestHandleGetIssues_FilterByOnHold(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	normal := seedIssue(db, p.GetId(), actor.GetId(), 1, models.IssueStatusTodo, true)
	held := seedIssue(db, p.GetId(), actor.GetId(), 2, models.IssueStatusInProgress, true)
	reason := models.HoldReasonWaitingOnExternal
	now := held.GetUpdatedAt()
	held.SetHold(true, &reason, &now, nil)
	db.Issues().Update(held)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	// Filter on_hold=true
	onHold := true
	result, err := queries.HandleGetIssues(ctx, queries.GetIssuesQuery{
		ProjectSlug: "p",
		OnHold:      &onHold,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 2, result.Items[0].Number)

	// Filter on_hold=false
	notOnHold := false
	result, err = queries.HandleGetIssues(ctx, queries.GetIssuesQuery{
		ProjectSlug: "p",
		OnHold:      &notOnHold,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 1, result.Items[0].Number)

	_ = normal // used above
}

func TestHandleGetIssues_ExcludeLabelID(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	labelID := uuid.New()

	labeled := seedIssue(db, p.GetId(), actor.GetId(), 1, models.IssueStatusTodo, true)
	labeled.SetLabels([]uuid.UUID{labelID})
	db.Issues().Update(labeled)
	seedIssue(db, p.GetId(), actor.GetId(), 2, models.IssueStatusTodo, true)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)

	// Exclude by label
	result, err := queries.HandleGetIssues(ctx, queries.GetIssuesQuery{
		ProjectSlug:    "p",
		ExcludeLabelID: &labelID,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 2, result.Items[0].Number)
}
