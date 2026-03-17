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

func TestHandleTriageIssue_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))
	project := &models.Project{ID: uuid.New(), Slug: "p", Name: "P", Archetype: models.ProjectArchetypeSoftware, CreatedBy: actor.ID}
	require.NoError(t, db.Projects().Insert(context.Background(), project))

	reporterID := actor.ID
	untriaged := &models.Issue{
		ID:         uuid.New(),
		ProjectID:  project.ID,
		Number:     1,
		Type:       models.IssueTypeTask,
		Title:      "Incoming",
		Status:     models.IssueStatusTodo,
		Priority:   models.IssuePriorityNone,
		Estimate:   models.IssueEstimateNone,
		ReporterID: &reporterID,
		Triaged:    false,
		Visibility: models.IssueVisibilityNormal,
		Checklist:  []models.ChecklistItem{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, db.Issues().Insert(context.Background(), untriaged))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleTriageIssue(ctx, commands.TriageIssueCommand{
		IssueID: untriaged.ID,
		Status:  models.IssueStatusInProgress,
	})
	require.NoError(t, err)

	issue, err := db.Issues().GetByID(context.Background(), untriaged.ID)
	require.NoError(t, err)
	assert.True(t, issue.Triaged)
	assert.Equal(t, models.IssueStatusInProgress, issue.Status)
}
