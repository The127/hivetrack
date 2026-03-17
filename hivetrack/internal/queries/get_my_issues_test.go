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

func TestHandleGetMyIssues_ReturnsAssignedNonTerminal(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))

	p := &models.Project{ID: uuid.New(), Slug: "p", Name: "P", Archetype: models.ProjectArchetypeSoftware, CreatedBy: actor.ID}
	require.NoError(t, db.Projects().Insert(context.Background(), p))

	i1 := seedIssue(db, p.ID, actor.ID, 1, models.IssueStatusTodo, true)
	i1.Assignees = []uuid.UUID{actor.ID}
	require.NoError(t, db.Issues().Update(context.Background(), i1))

	// This one is done (terminal) — should NOT appear
	i2 := seedIssue(db, p.ID, actor.ID, 2, models.IssueStatusDone, true)
	i2.Assignees = []uuid.UUID{actor.ID}
	require.NoError(t, db.Issues().Update(context.Background(), i2))

	// This one is not assigned to actor — should NOT appear
	_ = seedIssue(db, p.ID, actor.ID, 3, models.IssueStatusTodo, true)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetMyIssues(ctx, queries.GetMyIssuesQuery{})
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, i1.ID, result.Items[0].ID)
}
