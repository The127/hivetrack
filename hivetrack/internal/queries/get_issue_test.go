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

func TestHandleGetIssue_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))

	p := &models.Project{ID: uuid.New(), Slug: "myproj", Name: "P", Archetype: models.ProjectArchetypeSoftware, CreatedBy: actor.ID}
	require.NoError(t, db.Projects().Insert(context.Background(), p))

	issue := seedIssue(db, p.ID, actor.ID, 1, models.IssueStatusTodo, true)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetIssue(ctx, queries.GetIssueQuery{
		ProjectSlug: "myproj",
		Number:      issue.Number,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, issue.ID, result.ID)
	assert.Equal(t, issue.Title, result.Title)
}

func TestHandleGetIssue_NotFound(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))

	p := &models.Project{ID: uuid.New(), Slug: "myproj", Name: "P", Archetype: models.ProjectArchetypeSoftware, CreatedBy: actor.ID}
	require.NoError(t, db.Projects().Insert(context.Background(), p))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := queries.HandleGetIssue(ctx, queries.GetIssueQuery{
		ProjectSlug: "myproj",
		Number:      999,
	})
	require.NoError(t, err)
	assert.Nil(t, result)
}
