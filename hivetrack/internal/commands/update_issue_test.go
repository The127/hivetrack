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

func newTestIssue(projectID uuid.UUID, reporterID uuid.UUID, number int) *models.Issue {
	return &models.Issue{
		ID:        uuid.New(),
		ProjectID: projectID,
		Number:    number,
		Type:      models.IssueTypeTask,
		Title:     "Original title",
		Status:    models.IssueStatusTodo,
		Priority:  models.IssuePriorityNone,
		Estimate:  models.IssueEstimateNone,
		ReporterID: func() *uuid.UUID { id := reporterID; return &id }(),
		Triaged:    true,
		Visibility: models.IssueVisibilityNormal,
		Checklist:  []models.ChecklistItem{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func TestHandleUpdateIssue_ChangeTitle(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))
	project := &models.Project{ID: uuid.New(), Slug: "p", Name: "P", Archetype: models.ProjectArchetypeSoftware, CreatedBy: actor.ID}
	require.NoError(t, db.Projects().Insert(context.Background(), project))

	issue := newTestIssue(project.ID, actor.ID, 1)
	require.NoError(t, db.Issues().Insert(context.Background(), issue))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	newTitle := "New title"
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID: issue.ID,
		Title:   &newTitle,
	})
	require.NoError(t, err)

	updated, err := db.Issues().GetByID(context.Background(), issue.ID)
	require.NoError(t, err)
	assert.Equal(t, "New title", updated.Title)
}

func TestHandleUpdateIssue_SetOnHold(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))
	project := &models.Project{ID: uuid.New(), Slug: "p", Name: "P", Archetype: models.ProjectArchetypeSoftware, CreatedBy: actor.ID}
	require.NoError(t, db.Projects().Insert(context.Background(), project))

	issue := newTestIssue(project.ID, actor.ID, 1)
	require.NoError(t, db.Issues().Insert(context.Background(), issue))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	onHold := true
	reason := models.HoldReasonWaitingOnCustomer
	note := "waiting for response"
	_, err := commands.HandleUpdateIssue(ctx, commands.UpdateIssueCommand{
		IssueID:    issue.ID,
		OnHold:     &onHold,
		HoldReason: &reason,
		HoldNote:   &note,
	})
	require.NoError(t, err)

	updated, err := db.Issues().GetByID(context.Background(), issue.ID)
	require.NoError(t, err)
	assert.True(t, updated.OnHold)
	require.NotNil(t, updated.HoldReason)
	assert.Equal(t, models.HoldReasonWaitingOnCustomer, *updated.HoldReason)
}
