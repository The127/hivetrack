package commands_test

import (
	"context"
	"testing"

	"github.com/The127/mediatr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/events"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

func newTestIssue(projectID uuid.UUID, reporterID uuid.UUID, number int) *models.Issue {
	return models.NewIssue(
		projectID, number, models.IssueTypeTask, "Original title",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
}

func TestHandleUpdateIssue_AddLabels(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	label := models.NewLabel(project.GetId(), "bug", "#ff0000")
	db.Labels().Insert(label)
	issue := newTestIssue(project.GetId(), actor.GetId(), 1)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID:  issue.GetId(),
		LabelIDs: []uuid.UUID{label.GetId()},
	})
	require.NoError(t, err)

	updated, err := db.Issues().GetByID(context.Background(), issue.GetId())
	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{label.GetId()}, updated.GetLabels())
}

func TestHandleUpdateIssue_ClearLabels(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	label := models.NewLabel(project.GetId(), "bug", "#ff0000")
	db.Labels().Insert(label)
	issue := newTestIssue(project.GetId(), actor.GetId(), 1)
	issue.SetLabels([]uuid.UUID{label.GetId()})
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID:  issue.GetId(),
		LabelIDs: []uuid.UUID{},
	})
	require.NoError(t, err)

	updated, err := db.Issues().GetByID(context.Background(), issue.GetId())
	require.NoError(t, err)
	assert.Empty(t, updated.GetLabels())
}

func TestHandleUpdateIssue_ChangeTitle(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	issue := newTestIssue(project.GetId(), actor.GetId(), 1)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	newTitle := "New title"
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID: issue.GetId(),
		Title:   &newTitle,
	})
	require.NoError(t, err)

	updated, err := db.Issues().GetByID(context.Background(), issue.GetId())
	require.NoError(t, err)
	assert.Equal(t, "New title", updated.GetTitle())
}

func newTestEpic(projectID uuid.UUID, reporterID uuid.UUID, number int) *models.Issue {
	return models.NewIssue(
		projectID, number, models.IssueTypeEpic, "Test Epic",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
}

func TestHandleUpdateIssue_AssignParent(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	epic := newTestEpic(project.GetId(), actor.GetId(), 1)
	db.Issues().Insert(epic)
	task := newTestIssue(project.GetId(), actor.GetId(), 2)
	db.Issues().Insert(task)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	epicID := epic.GetId()
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID:  task.GetId(),
		ParentID: &epicID,
	})
	require.NoError(t, err)

	updated, err := db.Issues().GetByID(context.Background(), task.GetId())
	require.NoError(t, err)
	require.NotNil(t, updated.GetParentID())
	assert.Equal(t, epic.GetId(), *updated.GetParentID())
}

func TestHandleUpdateIssue_ClearParent(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	epic := newTestEpic(project.GetId(), actor.GetId(), 1)
	db.Issues().Insert(epic)
	task := newTestIssue(project.GetId(), actor.GetId(), 2)
	db.Issues().Insert(task)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	epicID := epic.GetId()
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID:  task.GetId(),
		ParentID: &epicID,
	})
	require.NoError(t, err)

	_, err = commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID:       task.GetId(),
		ClearParentID: true,
	})
	require.NoError(t, err)

	updated, err := db.Issues().GetByID(context.Background(), task.GetId())
	require.NoError(t, err)
	assert.Nil(t, updated.GetParentID())
}

func TestHandleUpdateIssue_EpicCantHaveParent(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	epic1 := newTestEpic(project.GetId(), actor.GetId(), 1)
	epic2 := newTestEpic(project.GetId(), actor.GetId(), 2)
	db.Issues().Insert(epic1)
	db.Issues().Insert(epic2)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	parentID := epic1.GetId()
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID:  epic2.GetId(),
		ParentID: &parentID,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, models.ErrBadRequest)
}

func TestHandleUpdateIssue_ParentMustBeEpic(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	task1 := newTestIssue(project.GetId(), actor.GetId(), 1)
	task2 := newTestIssue(project.GetId(), actor.GetId(), 2)
	db.Issues().Insert(task1)
	db.Issues().Insert(task2)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	parentID := task1.GetId()
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID:  task2.GetId(),
		ParentID: &parentID,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, models.ErrBadRequest)
}

func TestHandleUpdateIssue_ParentMustBeSameProject(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project1 := models.NewProject(actor.GetId(), "p1", "P1", models.ProjectArchetypeSoftware)
	project2 := models.NewProject(actor.GetId(), "p2", "P2", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project1)
	db.Projects().Insert(project2)
	require.NoError(t, db.SaveChanges(context.Background()))

	epic := newTestEpic(project1.GetId(), actor.GetId(), 1)
	task := newTestIssue(project2.GetId(), actor.GetId(), 1)
	db.Issues().Insert(epic)
	db.Issues().Insert(task)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	epicID := epic.GetId()
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID:  task.GetId(),
		ParentID: &epicID,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, models.ErrBadRequest)
}

func TestHandleUpdateIssue_TodoToInProgress_AutoAssignsActor(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	issue := newTestIssue(project.GetId(), actor.GetId(), 1)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	m := mediatr.NewMediator()
	mediatr.RegisterEventHandler(m, events.HandleAutoAssignOnStatusChange)

	ctx := commands.ContextWithMediator(testutil.ContextWithUser(testutil.ContextWithDb(db), actor), m)
	newStatus := models.IssueStatusInProgress
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID: issue.GetId(),
		Status:  &newStatus,
	})
	require.NoError(t, err)

	updated, err := db.Issues().GetByID(context.Background(), issue.GetId())
	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{actor.GetId()}, updated.GetAssignees())
}

func TestHandleUpdateIssue_TodoToInProgress_AlreadyAssigned_NoChange(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	issue := newTestIssue(project.GetId(), actor.GetId(), 1)
	existingAssignee := uuid.New()
	issue.SetAssignees([]uuid.UUID{existingAssignee})
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	m := mediatr.NewMediator()
	mediatr.RegisterEventHandler(m, events.HandleAutoAssignOnStatusChange)

	ctx := commands.ContextWithMediator(testutil.ContextWithUser(testutil.ContextWithDb(db), actor), m)
	newStatus := models.IssueStatusInProgress
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID: issue.GetId(),
		Status:  &newStatus,
	})
	require.NoError(t, err)

	updated, err := db.Issues().GetByID(context.Background(), issue.GetId())
	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{existingAssignee}, updated.GetAssignees())
}

func TestHandleUpdateIssue_MarkTaskAsRefined(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	issue := newTestIssue(project.GetId(), actor.GetId(), 1)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	require.False(t, issue.GetRefined(), "precondition: issue must start as unrefined")

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	refined := true
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID: issue.GetId(),
		Refined: &refined,
	})
	require.NoError(t, err, "HandleUpdateIssue should succeed when marking a task as refined")

	updated, err := db.Issues().GetByID(context.Background(), issue.GetId())
	require.NoError(t, err, "should be able to retrieve the issue after update")
	assert.True(t, updated.GetRefined(), "issue should be marked as refined after update")
}

func TestHandleUpdateIssue_IssueRefinedEventEnqueuedInOutboxWhenRefined(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	issue := newTestIssue(project.GetId(), actor.GetId(), 1)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	require.False(t, issue.GetRefined(), "precondition: issue must start as unrefined")

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	refined := true
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID: issue.GetId(),
		Refined: &refined,
	})
	require.NoError(t, err)

	pending, err := db.Outbox().ListPending(context.Background())
	require.NoError(t, err)
	require.Len(t, pending, 1, "expected exactly 1 outbox message after refining an issue")
	assert.Equal(t, events.EventTypeIssueRefined, pending[0].Type, "outbox message type should be %q", events.EventTypeIssueRefined)
	assert.Contains(t, string(pending[0].Payload), issue.GetId().String(), "outbox payload should reference the refined issue ID")
}

func TestHandleUpdateIssue_SetOnHold(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	issue := newTestIssue(project.GetId(), actor.GetId(), 1)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	onHold := true
	reason := models.HoldReasonWaitingOnCustomer
	note := "waiting for response"
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID:    issue.GetId(),
		OnHold:     &onHold,
		HoldReason: &reason,
		HoldNote:   &note,
	})
	require.NoError(t, err)

	updated, err := db.Issues().GetByID(context.Background(), issue.GetId())
	require.NoError(t, err)
	assert.True(t, updated.GetOnHold())
	require.NotNil(t, updated.GetHoldReason())
	assert.Equal(t, models.HoldReasonWaitingOnCustomer, *updated.GetHoldReason())
}

func TestHandleUpdateIssue_RefinedRejectedForEpic(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	epic := newTestEpic(project.GetId(), actor.GetId(), 1)
	db.Issues().Insert(epic)
	require.NoError(t, db.SaveChanges(context.Background()))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	refined := true
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID: epic.GetId(),
		Refined: &refined,
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, models.ErrBadRequest)

	var domainErr *models.DomainError
	require.ErrorAs(t, err, &domainErr, "expected a DomainError with a specific code")
	assert.Equal(t, "refined_not_supported_for_epics", domainErr.Code)

	// Verify the issue was not mutated.
	unchanged, err := db.Issues().GetByID(context.Background(), epic.GetId())
	require.NoError(t, err)
	assert.False(t, unchanged.GetRefined())
}
