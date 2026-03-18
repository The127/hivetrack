//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
	"github.com/the127/hivetrack/internal/repositories/postgres"
)

func setupProjectWithUser(t *testing.T, ctx context.Context, db *postgres.DbContext) (*models.User, *models.Project) {
	t.Helper()
	user := newTestUser(t, ctx, db)
	project := newTestProject(t, user.GetId())
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(ctx))
	return user, project
}

func newTestIssue(t *testing.T, ctx context.Context, db *postgres.DbContext, projectID uuid.UUID, reporterID uuid.UUID) *models.Issue {
	t.Helper()
	num, err := db.Projects().NextIssueNumber(ctx, projectID)
	require.NoError(t, err)
	return models.NewIssue(
		projectID, num,
		models.IssueTypeTask, "Test Issue",
		models.IssueStatusTodo, models.IssuePriorityMedium, models.IssueEstimateS,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil,
		nil, nil,
	)
}

func TestIssueRepository_InsertAndGetByID(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user, project := setupProjectWithUser(t, ctx, db)
	issue := newTestIssue(t, ctx, db, project.GetId(), user.GetId())

	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(ctx))

	got, err := db.Issues().GetByID(ctx, issue.GetId())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Test Issue", got.GetTitle())
	assert.Equal(t, models.IssueStatusTodo, got.GetStatus())
	assert.Equal(t, models.IssuePriorityMedium, got.GetPriority())
	assert.Equal(t, models.IssueEstimateS, got.GetEstimate())
	assert.True(t, got.GetTriaged())
}

func TestIssueRepository_GetByNumber(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user, project := setupProjectWithUser(t, ctx, db)
	issue := newTestIssue(t, ctx, db, project.GetId(), user.GetId())

	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(ctx))

	got, err := db.Issues().GetByNumber(ctx, project.GetId(), issue.GetNumber())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, issue.GetId(), got.GetId())
}

func TestIssueRepository_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	got, err := db.Issues().GetByID(ctx, uuid.New())
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestIssueRepository_Update(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user, project := setupProjectWithUser(t, ctx, db)
	issue := newTestIssue(t, ctx, db, project.GetId(), user.GetId())
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(ctx))

	issue.SetTitle("Updated Title")
	issue.SetStatus(models.IssueStatusInProgress)
	issue.SetPriority(models.IssuePriorityHigh)
	db.Issues().Update(issue)
	require.NoError(t, db.SaveChanges(ctx))

	got, err := db.Issues().GetByID(ctx, issue.GetId())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Updated Title", got.GetTitle())
	assert.Equal(t, models.IssueStatusInProgress, got.GetStatus())
	assert.Equal(t, models.IssuePriorityHigh, got.GetPriority())
}

func TestIssueRepository_Delete(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user, project := setupProjectWithUser(t, ctx, db)
	issue := newTestIssue(t, ctx, db, project.GetId(), user.GetId())
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(ctx))

	db.Issues().Delete(issue)
	require.NoError(t, db.SaveChanges(ctx))

	got, err := db.Issues().GetByID(ctx, issue.GetId())
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestIssueRepository_List_ByProject(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user, project := setupProjectWithUser(t, ctx, db)
	i1 := newTestIssue(t, ctx, db, project.GetId(), user.GetId())
	i2 := newTestIssue(t, ctx, db, project.GetId(), user.GetId())
	db.Issues().Insert(i1)
	db.Issues().Insert(i2)
	require.NoError(t, db.SaveChanges(ctx))

	issues, total, err := db.Issues().List(ctx, repositories.NewIssueFilter().ByProjectID(project.GetId()))
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, issues, 2)
}

func TestIssueRepository_List_ByStatus(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user, project := setupProjectWithUser(t, ctx, db)
	reporterID := user.GetId()

	todo := newTestIssue(t, ctx, db, project.GetId(), reporterID)

	doneNum, err := db.Projects().NextIssueNumber(ctx, project.GetId())
	require.NoError(t, err)
	done := models.NewIssue(project.GetId(), doneNum,
		models.IssueTypeTask, "Done Issue",
		models.IssueStatusDone, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil)

	db.Issues().Insert(todo)
	db.Issues().Insert(done)
	require.NoError(t, db.SaveChanges(ctx))

	issues, total, err := db.Issues().List(ctx,
		repositories.NewIssueFilter().ByProjectID(project.GetId()).ByStatus(models.IssueStatusTodo))
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, issues, 1)
	assert.Equal(t, todo.GetId(), issues[0].GetId())
}

func TestIssueRepository_List_ByAssignee(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user, project := setupProjectWithUser(t, ctx, db)
	other := newTestUser(t, ctx, db)

	userID := user.GetId()
	otherID := other.GetId()

	assigned := newTestIssue(t, ctx, db, project.GetId(), user.GetId())
	assigned.SetAssignees([]uuid.UUID{userID})

	unassigned := newTestIssue(t, ctx, db, project.GetId(), user.GetId())
	unassigned.SetAssignees([]uuid.UUID{otherID})

	db.Issues().Insert(assigned)
	db.Issues().Insert(unassigned)
	require.NoError(t, db.SaveChanges(ctx))

	issues, total, err := db.Issues().List(ctx,
		repositories.NewIssueFilter().ByProjectID(project.GetId()).ByAssigneeID(userID))
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, issues, 1)
	assert.Equal(t, assigned.GetId(), issues[0].GetId())
}

func TestIssueRepository_List_Triaged(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user, project := setupProjectWithUser(t, ctx, db)
	reporterID := user.GetId()

	triagedNum, err := db.Projects().NextIssueNumber(ctx, project.GetId())
	require.NoError(t, err)
	triaged := models.NewIssue(project.GetId(), triagedNum,
		models.IssueTypeTask, "Triaged",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil)

	inboxNum, err := db.Projects().NextIssueNumber(ctx, project.GetId())
	require.NoError(t, err)
	inbox := models.NewIssue(project.GetId(), inboxNum,
		models.IssueTypeTask, "Inbox",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, false, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil)

	db.Issues().Insert(triaged)
	db.Issues().Insert(inbox)
	require.NoError(t, db.SaveChanges(ctx))

	issues, total, err := db.Issues().List(ctx,
		repositories.NewIssueFilter().ByProjectID(project.GetId()).WithTriaged(false))
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, issues, 1)
	assert.Equal(t, inbox.GetId(), issues[0].GetId())
}

func TestIssueRepository_WithAssigneesAndLabels(t *testing.T) {
	ctx := context.Background()
	db := postgres.NewDbContext(testDB)

	user, project := setupProjectWithUser(t, ctx, db)

	label := models.NewLabel(project.GetId(), "bug", "#ff0000")
	db.Labels().Insert(label)
	require.NoError(t, db.SaveChanges(ctx))

	issue := newTestIssue(t, ctx, db, project.GetId(), user.GetId())
	issue.SetAssignees([]uuid.UUID{user.GetId()})
	issue.SetLabels([]uuid.UUID{label.GetId()})
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(ctx))

	got, err := db.Issues().GetByID(ctx, issue.GetId())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, []uuid.UUID{user.GetId()}, got.GetAssignees())
	assert.Equal(t, []uuid.UUID{label.GetId()}, got.GetLabels())
}
