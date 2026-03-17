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
