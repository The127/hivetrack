package setup

import (
	"context"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/infrastructure"
)

// refinementPublisherAdapter bridges commands.RefinementPublisher to infrastructure.NatsPublisher.
type refinementPublisherAdapter struct {
	pub *infrastructure.NatsPublisher
}

func (a *refinementPublisherAdapter) PublishRefinementRequest(ctx context.Context, req commands.RefinementPublishRequest) error {
	msgs := make([]infrastructure.RefinementChatMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = infrastructure.RefinementChatMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}
	return a.pub.PublishRefinementRequest(ctx, infrastructure.RefinementRequest{
		SessionID:   req.SessionID,
		IssueID:     req.IssueID,
		ProjectSlug: req.ProjectSlug,
		Title:       req.Title,
		Description: req.Description,
		Phase:       req.Phase,
		Messages:    msgs,
	})
}

func (a *refinementPublisherAdapter) PublishRefinementAccept(ctx context.Context, sessionID uuid.UUID) error {
	return a.pub.PublishRefinementAccept(ctx, sessionID)
}

func (a *refinementPublisherAdapter) PublishStoryRefined(ctx context.Context, event commands.StoryRefinedEvent) error {
	return a.pub.PublishStoryRefined(ctx, infrastructure.StoryRefinedEvent{
		StoryID:             event.StoryID,
		ProjectID:           event.ProjectID,
		ProjectSlug:         event.ProjectSlug,
		IssueNumber:         event.IssueNumber,
		Title:               event.Title,
		Actor:               event.Actor,
		Goal:                event.Goal,
		MainSuccessScenario: event.MainSuccessScenario,
		Preconditions:       event.Preconditions,
		AcceptanceCriteria:  event.AcceptanceCriteria,
		Extensions:          event.Extensions,
	})
}
