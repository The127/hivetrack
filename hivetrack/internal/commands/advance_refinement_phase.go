package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type AdvanceRefinementPhaseCommand struct {
	IssueID     uuid.UUID
	TargetPhase string // optional: if empty, advance to next; if set, jump to that phase
}

type AdvanceRefinementPhaseResult struct {
	Phase string `json:"phase"`
}

func NewAdvanceRefinementPhaseHandler(publisher RefinementPublisher) func(context.Context, AdvanceRefinementPhaseCommand) (*AdvanceRefinementPhaseResult, error) {
	return func(ctx context.Context, cmd AdvanceRefinementPhaseCommand) (*AdvanceRefinementPhaseResult, error) {
		db := repositories.GetDbContext(ctx)

		// Load active session
		session, err := db.Refinements().GetActiveSession(ctx, cmd.IssueID)
		if err != nil {
			return nil, fmt.Errorf("getting active session: %w", err)
		}
		if session == nil {
			return nil, fmt.Errorf("no active refinement session for issue %s: %w", cmd.IssueID, models.ErrNotFound)
		}

		// Determine target phase
		var newPhase models.RefinementPhase
		if cmd.TargetPhase != "" {
			if !models.ValidPhase(cmd.TargetPhase) {
				return nil, fmt.Errorf("invalid phase %q: %w", cmd.TargetPhase, models.ErrBadRequest)
			}
			newPhase = models.RefinementPhase(cmd.TargetPhase)
		} else {
			next, err := models.NextPhase(session.CurrentPhase)
			if err != nil {
				return nil, err
			}
			newPhase = next
		}

		// Update session phase
		if err := db.Refinements().UpdateSessionPhase(ctx, session.ID, newPhase); err != nil {
			return nil, fmt.Errorf("updating session phase: %w", err)
		}

		// Load full message history
		_, messages, err := db.Refinements().GetSessionWithMessages(ctx, session.ID)
		if err != nil {
			return nil, fmt.Errorf("loading message history: %w", err)
		}

		// Load issue and project
		issue, err := db.Issues().GetByID(ctx, cmd.IssueID)
		if err != nil {
			return nil, fmt.Errorf("getting issue: %w", err)
		}

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

		// Publish to NATS with new phase
		if err := publisher.PublishRefinementRequest(ctx, RefinementPublishRequest{
			SessionID:   session.ID,
			IssueID:     cmd.IssueID,
			ProjectSlug: project.GetSlug(),
			Title:       issue.GetTitle(),
			Description: issue.GetDescription(),
			Phase:       string(newPhase),
			Messages:    chatMessages,
		}); err != nil {
			return nil, fmt.Errorf("publishing refinement request: %w", err)
		}

		return &AdvanceRefinementPhaseResult{Phase: string(newPhase)}, nil
	}
}
