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

func TestSendRefinementMessage_Success(t *testing.T) {
	db := inmemory.NewDbContext()
	actor := models.NewUser("sub1", "test@example.com", "test@example.com")
	require.NoError(t, db.Users().Upsert(context.Background(), actor))
	project := models.NewProject(actor.GetId(), "proj", "Proj", models.ProjectArchetypeSoftware)
	db.Projects().Insert(project)
	require.NoError(t, db.SaveChanges(context.Background()))

	reporterID := actor.GetId()
	desc := "Original desc"
	issue := models.NewIssue(
		project.GetId(), 1, models.IssueTypeTask, "The task",
		models.IssueStatusTodo, models.IssuePriorityNone, models.IssueEstimateNone,
		&reporterID, true, models.IssueVisibilityNormal,
		&desc, nil, nil, nil, nil,
	)
	db.Issues().Insert(issue)
	require.NoError(t, db.SaveChanges(context.Background()))

	// Create active session
	session := models.NewRefinementSession(issue.GetId())
	require.NoError(t, db.Refinements().CreateSession(context.Background(), session))

	pub := &spyPublisher{}
	handler := commands.NewSendRefinementMessageHandler(pub, func(uuid.UUID) {})

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := handler(ctx, commands.SendRefinementMessageCommand{
		IssueID: issue.GetId(),
		Content: "What about edge cases?",
	})
	require.NoError(t, err)

	// Message was stored
	_, msgs, err := db.Refinements().GetSessionWithMessages(context.Background(), session.ID)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	assert.Equal(t, "What about edge cases?", msgs[0].Content)
	assert.Equal(t, models.RefinementRoleUser, msgs[0].Role)
	assert.Equal(t, models.RefinementPhaseActorGoal, msgs[0].Phase)

	// NATS request was published with message history
	require.Len(t, pub.published, 1)
	req := pub.published[0]
	assert.Equal(t, session.ID, req.SessionID)
	assert.Equal(t, "proj", req.ProjectSlug)
	require.Len(t, req.Messages, 1)
	assert.Equal(t, "user", req.Messages[0].Role)
	assert.Equal(t, "What about edge cases?", req.Messages[0].Content)
	assert.Equal(t, "actor_goal", req.Phase)
}

func TestSendRefinementMessage_NoActiveSession(t *testing.T) {
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

	pub := &spyPublisher{}
	handler := commands.NewSendRefinementMessageHandler(pub, func(uuid.UUID) {})

	ctx := testutil.ContextWithUser(testutil.ContextWithDb(db), actor)
	_, err := handler(ctx, commands.SendRefinementMessageCommand{
		IssueID: issue.GetId(),
		Content: "Hello",
	})
	require.ErrorIs(t, err, models.ErrNotFound)
	assert.Empty(t, pub.published)
}
