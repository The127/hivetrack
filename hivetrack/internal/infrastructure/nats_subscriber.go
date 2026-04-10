package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

const SubjectRefinementResponse = "hivemind-refinement.response"

// RefinementResponse is the message received from Hivemind via NATS.
type RefinementResponse struct {
	SessionID   uuid.UUID              `json:"session_id"`
	IssueID     uuid.UUID              `json:"issue_id"`
	Phase       string                 `json:"phase"`
	Type        string                 `json:"type"` // "question", "proposal", or "phase_result"
	Content     string                 `json:"content"`
	Proposal    *RefinementProposal    `json:"proposal"`
	PhaseData   map[string]interface{} `json:"phase_data"`
	Suggestions []string               `json:"suggestions"`
}

// RefinementProposal is the proposed title/description from Hivemind.
type RefinementProposal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// NatsSubscriber listens for Hivemind refinement responses and stores them.
type NatsSubscriber struct {
	js          jetstream.JetStream
	newRepo     func() repositories.RefinementRepository
	logger      *zap.Logger
	tokenBuffer *TokenBuffer // may be nil
}

// NewNatsSubscriber creates a subscriber. newRepo is called per message to get a fresh repository.
func NewNatsSubscriber(js jetstream.JetStream, newRepo func() repositories.RefinementRepository, logger *zap.Logger, buf *TokenBuffer) *NatsSubscriber {
	return &NatsSubscriber{
		js:          js,
		newRepo:     newRepo,
		logger:      logger,
		tokenBuffer: buf,
	}
}

// Start begins consuming refinement responses in the background.
// Cancel the context to stop the subscriber.
func (s *NatsSubscriber) Start(ctx context.Context) error {
	consumer, err := s.js.CreateOrUpdateConsumer(ctx, "hivemind-refinement", jetstream.ConsumerConfig{
		Durable:       "hivetrack-refinement-consumer",
		FilterSubject: SubjectRefinementResponse,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return fmt.Errorf("creating NATS consumer: %w", err)
	}

	go func() {
		for {
			msgs, err := consumer.Fetch(1, jetstream.FetchMaxWait(5*time.Second))
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				s.logger.Error("fetching NATS message", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			for msg := range msgs.Messages() {
				if err := s.handleMessage(ctx, msg); err != nil {
					s.logger.Error("handling refinement response", zap.Error(err))
					// ACK rather than NAK — stale messages (e.g. deleted sessions)
					// would loop forever on NAK. Transient errors are rare enough
					// that discarding is acceptable.
					if ackErr := msg.Ack(); ackErr != nil {
						s.logger.Error("acking failed message", zap.Error(ackErr))
					}
				} else {
					if ackErr := msg.Ack(); ackErr != nil {
						s.logger.Error("acking message", zap.Error(ackErr))
					}
				}
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}

func (s *NatsSubscriber) handleMessage(ctx context.Context, msg jetstream.Msg) error {
	var resp RefinementResponse
	if err := json.Unmarshal(msg.Data(), &resp); err != nil {
		return fmt.Errorf("unmarshaling refinement response: %w", err)
	}

	msgType := models.RefinementMessageTypeMessage
	var proposal *models.RefinementProposal
	switch resp.Type {
	case "proposal":
		if resp.Proposal != nil {
			msgType = models.RefinementMessageTypeProposal
			proposal = &models.RefinementProposal{
				Title:       resp.Proposal.Title,
				Description: resp.Proposal.Description,
			}
		}
	case "phase_result":
		msgType = models.RefinementMessageTypePhaseResult
	}

	phase := models.RefinementPhase(resp.Phase)
	if !models.ValidPhase(resp.Phase) {
		phase = models.RefinementPhaseActorGoal
	}

	refinementMsg := models.NewRefinementMessage(
		resp.SessionID,
		models.RefinementRoleAssistant,
		resp.Content,
		msgType,
		phase,
		proposal,
	)
	refinementMsg.PhaseData = resp.PhaseData
	refinementMsg.Suggestions = resp.Suggestions

	repo := s.newRepo()
	if err := repo.AddMessage(ctx, refinementMsg); err != nil {
		return fmt.Errorf("storing refinement response: %w", err)
	}

	if s.tokenBuffer != nil {
		s.tokenBuffer.ClearPartialResponse(resp.SessionID)
	}

	s.logger.Info("stored refinement response",
		zap.String("session_id", resp.SessionID.String()),
		zap.String("type", resp.Type),
	)
	return nil
}
