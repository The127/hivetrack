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
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	p := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(p)
	require.NoError(t, db.SaveChanges(context.Background()))

	i1 := seedIssue(db, p.GetId(), actor.GetId(), 1, models.IssueStatusTodo, true)
	i1.SetAssignees([]uuid.UUID{actor.GetId()})
	db.Issues().Update(i1)
	require.NoError(t, db.SaveChanges(context.Background()))

	// This one is done (terminal) — should NOT appear
	i2 := seedIssue(db, p.GetId(), actor.GetId(), 2, models.IssueStatusDone, true)
	i2.SetAssignees([]uuid.UUID{actor.GetId()})
	db.Issues().Update(i2)
	require.NoError(t, db.SaveChanges(context.Background()))

	// This one is not assigned to actor — should NOT appear
	_ = seedIssue(db, p.GetId(), actor.GetId(), 3, models.IssueStatusTodo, true)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetMyIssues(ctx, queries.GetMyIssuesQuery{})
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, i1.GetId(), result.Items[0].ID)
}
