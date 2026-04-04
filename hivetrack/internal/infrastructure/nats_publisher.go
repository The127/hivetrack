package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
)

const SubjectRefinementRequest = "hivetrack-refinement.request"

// RefinementRequest is the message published to NATS for Hivemind to process.
type RefinementRequest struct {
	SessionID   uuid.UUID               `json:"session_id"`
	IssueID     uuid.UUID               `json:"issue_id"`
	ProjectSlug string                  `json:"project_slug"`
	Title       string                  `json:"title"`
	Description *string                 `json:"description"`
	Messages    []RefinementChatMessage `json:"messages"`
}

// RefinementChatMessage is a single message in the refinement conversation.
type RefinementChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NatsPublisher publishes refinement requests to NATS JetStream.
type NatsPublisher struct {
	js jetstream.JetStream
}

func NewNatsPublisher(js jetstream.JetStream) *NatsPublisher {
	return &NatsPublisher{js: js}
}

func (p *NatsPublisher) PublishRefinementRequest(ctx context.Context, req RefinementRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshaling refinement request: %w", err)
	}

	if _, err := p.js.Publish(ctx, SubjectRefinementRequest, data); err != nil {
		return fmt.Errorf("publishing refinement request: %w", err)
	}
	return nil
}

// RefinementAccept signals to Hivemind that the user accepted a proposal.
type RefinementAccept struct {
	SessionID uuid.UUID `json:"session_id"`
	Action    string    `json:"action"`
}

func (p *NatsPublisher) PublishRefinementAccept(ctx context.Context, sessionID uuid.UUID) error {
	data, err := json.Marshal(RefinementAccept{
		SessionID: sessionID,
		Action:    "accept",
	})
	if err != nil {
		return fmt.Errorf("marshaling refinement accept: %w", err)
	}

	if _, err := p.js.Publish(ctx, SubjectRefinementRequest, data); err != nil {
		return fmt.Errorf("publishing refinement accept: %w", err)
	}
	return nil
}
