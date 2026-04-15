package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type SendRefinementMessageCommand struct {
	IssueID uuid.UUID
	Content string
}

type SendRefinementMessageResult struct{}

// NewSendRefinementMessageHandler returns a handler that depends on a RefinementPublisher.
// notify is invoked post-commit so real-time subscribers can refetch the session.
func NewSendRefinementMessageHandler(publisher RefinementPublisher, notify func(uuid.UUID)) func(context.Context, SendRefinementMessageCommand) (*SendRefinementMessageResult, error) {
	return func(ctx context.Context, cmd SendRefinementMessageCommand) (*SendRefinementMessageResult, error) {
		db := repositories.GetDbContext(ctx)

		// Load active session
		session, err := db.Refinements().GetActiveSession(ctx, cmd.IssueID)
		if err != nil {
			return nil, fmt.Errorf("getting active session: %w", err)
		}
		if session == nil {
			return nil, fmt.Errorf("no active refinement session for issue %s: %w", cmd.IssueID, models.ErrNotFound)
		}

		// Store user message
		msg := models.NewRefinementMessage(session.ID, models.RefinementRoleUser, cmd.Content, models.RefinementMessageTypeMessage, session.CurrentPhase, nil)
		if err := db.Refinements().AddMessage(ctx, msg); err != nil {
			return nil, fmt.Errorf("storing user message: %w", err)
		}
		notify(cmd.IssueID)

		// Load full message history
		_, messages, err := db.Refinements().GetSessionWithMessages(ctx, session.ID)
		if err != nil {
			return nil, fmt.Errorf("loading message history: %w", err)
		}

		// Load issue
		issue, err := db.Issues().GetByID(ctx, cmd.IssueID)
		if err != nil {
			return nil, fmt.Errorf("getting issue: %w", err)
		}

		// Load project for slug
		project, err := db.Projects().GetByID(ctx, issue.GetProjectID())
		if err != nil {
			return nil, fmt.Errorf("getting project: %w", err)
		}

		// Build chat history
		chatMessages := make([]RefinementChatMessage, len(messages))
		for i, m := range messages {
			chatMessages[i] = RefinementChatMessage{
				Role:    string(m.Role),
				Content: m.Content,
			}
		}

		// Publish to NATS with full history
		if err := publisher.PublishRefinementRequest(ctx, RefinementPublishRequest{
			SessionID:   session.ID,
			IssueID:     cmd.IssueID,
			ProjectSlug: project.GetSlug(),
			Title:       issue.GetTitle(),
			Description: issue.GetDescription(),
			Phase:       string(session.CurrentPhase),
			Messages:    chatMessages,
		}); err != nil {
			return nil, fmt.Errorf("publishing refinement request: %w", err)
		}

		return &SendRefinementMessageResult{}, nil
	}
}
