package commands_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories/inmemory"
	"github.com/the127/hivetrack/internal/testutil"
)

func TestHandleDeleteIssue_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.User{ID: uuid.New(), Sub: "sub1", Email: "test@example.com"}
	require.NoError(t, db.Users().Upsert(context.Background(), &actor))
	project := &models.Project{ID: uuid.New(), Slug: "p", Name: "P", Archetype: models.ProjectArchetypeSoftware, CreatedBy: actor.ID}
	require.NoError(t, db.Projects().Insert(context.Background(), project))

	issue := newTestIssue(project.ID, actor.ID, 1)
	require.NoError(t, db.Issues().Insert(context.Background(), issue))

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := commands.HandleDeleteIssue(ctx, commands.DeleteIssueCommand{IssueID: issue.ID})
	require.NoError(t, err)

	deleted, err := db.Issues().GetByID(context.Background(), issue.ID)
	require.NoError(t, err)
	assert.Nil(t, deleted)
}
