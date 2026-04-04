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

type spyPublisher struct {
	published []commands.RefinementPublishRequest
}

func (s *spyPublisher) PublishRefinementRequest(_ context.Context, req commands.RefinementPublishRequest) error {
	s.published = append(s.published, req)
	return nil
}

func (s *spyPublisher) PublishRefinementAccept(_ context.Context, _ uuid.UUID) error {
	return nil
}

func TestStartRefinementSession_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "myproject", "My Project", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	reporterID := actor.GetId()
	desc := "Some description"
	issue := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "Fix the bug",
		models.IssueStatusTodo, models.IssuePriorityHigh, models.IssueEstimateM,
		&reporterID, true, models.IssueVisibilityNormal,
		&desc, nil, nil, nil, nil,
	)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	pub := &spyPublisher{}
	handler := commands.NewStartRefinementSessionHandler(pub)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	result, err := handler(ctx, commands.StartRefinementSessionCommand{
		IssueID: issue.GetId(),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, result.SessionID)

	// Session was created
	session, err := db.Refinements().GetActiveSession(context.Background(), issue.GetId())
	require.NoError(t, err)
	require.NotNil(t, session)
	assert.Equal(t, result.SessionID, session.ID)

	// NATS request was published
	require.Len(t, pub.published, 1)
	req := pub.published[0]
	assert.Equal(t, result.SessionID, req.SessionID)
	assert.Equal(t, issue.GetId(), req.IssueID)
	assert.Equal(t, "myproject", req.ProjectSlug)
	assert.Equal(t, "Fix the bug", req.Title)
	assert.Equal(t, &desc, req.Description)
	assert.Nil(t, req.Messages)
}

func TestStartRefinementSession_ConflictWhenActiveExists(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "p", "P", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	reporterID := actor.GetId()
	issue := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "Task",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		nil, nil, nil, nil, nil,
	)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	// Create an active session
	session := models.NewRefinementSession(issue.GetId())
	require.NoError(t, db.Refinements().CreateSession(context.Background(), session))

	pub := &spyPublisher{}
	handler := commands.NewStartRefinementSessionHandler(pub)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := handler(ctx, commands.StartRefinementSessionCommand{
		IssueID: issue.GetId(),
	})
	require.ErrorIs(t, err, models.ErrConflict)
	assert.Empty(t, pub.published)
}

func TestStartRefinementSession_IssueNotFound(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))

	pub := &spyPublisher{}
	handler := commands.NewStartRefinementSessionHandler(pub)

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := handler(ctx, commands.StartRefinementSessionCommand{
		IssueID: models.NewBaseModel().GetId(),
	})
	require.ErrorIs(t, err, models.ErrNotFound)
}
