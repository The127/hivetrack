package setup

import (
	"context"

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
		Messages:    msgs,
	})
}
