package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

const (
	SubjectRefinementResponse = "hivemind-refinement.response"
	refinementStreamName      = "hivemind-refinement"
	refinementConsumerName    = "hivetrack-refinement-consumer"
)

// RefinementResponse is the message received from Hivemind via NATS.
type RefinementResponse struct {
	SessionID uuid.UUID              `json:"session_id"`
	IssueID   uuid.UUID              `json:"issue_id"`
	Phase     string                 `json:"phase"`
	Type      string                 `json:"type"` // "question", "proposal", or "phase_result"
	Content   string                 `json:"content"`
	Proposal  *RefinementProposal    `json:"proposal"`
	PhaseData map[string]interface{} `json:"phase_data"`
}

// RefinementProposal is the proposed title/description from Hivemind.
type RefinementProposal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// NatsSubscriber listens for Hivemind refinement responses and stores them.
type NatsSubscriber struct {
	js      jetstream.JetStream
	newRepo func() repositories.RefinementRepository
	notify  func(uuid.UUID)
	logger  *zap.Logger
}

// NewNatsSubscriber creates a subscriber. newRepo is called per message to get a fresh repository.
// notify is invoked with the issue ID after a message has been stored so real-time
// subscribers (e.g. SSE streams) can refetch the session.
func NewNatsSubscriber(js jetstream.JetStream, newRepo func() repositories.RefinementRepository, notify func(uuid.UUID), logger *zap.Logger) *NatsSubscriber {
	return &NatsSubscriber{
		js:      js,
		newRepo: newRepo,
		notify:  notify,
		logger:  logger,
	}
}

// Start begins consuming refinement responses in the background.
// Cancel the context to stop the subscriber.
func (s *NatsSubscriber) Start(ctx context.Context) error {
	consumer, err := s.ensureConsumer(ctx)
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
				// NATS may have lost the consumer (restart, data loss, manual delete).
				// Re-create it so the subscriber self-heals instead of looping on a
				// ghost reference forever.
				if errors.Is(err, jetstream.ErrConsumerNotFound) || errors.Is(err, jetstream.ErrStreamNotFound) {
					newConsumer, rerr := s.ensureConsumer(ctx)
					if rerr != nil {
						s.logger.Error("recreating NATS consumer", zap.Error(rerr))
					} else {
						s.logger.Info("recreated NATS consumer after loss")
						consumer = newConsumer
					}
				}
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

// ensureConsumer creates-or-updates the refinement consumer, re-creating the
// underlying stream first if NATS has lost it (e.g. after a restart with
// ephemeral storage). Both operations are idempotent.
func (s *NatsSubscriber) ensureConsumer(ctx context.Context) (jetstream.Consumer, error) {
	cfg := jetstream.ConsumerConfig{
		Durable:       refinementConsumerName,
		FilterSubject: SubjectRefinementResponse,
		AckPolicy:     jetstream.AckExplicitPolicy,
	}
	consumer, err := s.js.CreateOrUpdateConsumer(ctx, refinementStreamName, cfg)
	if err == nil {
		return consumer, nil
	}
	if !errors.Is(err, jetstream.ErrStreamNotFound) {
		return nil, err
	}
	if _, serr := s.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     refinementStreamName,
		Subjects: []string{refinementStreamName + ".>"},
	}); serr != nil {
		return nil, fmt.Errorf("recreating stream: %w", serr)
	}
	return s.js.CreateOrUpdateConsumer(ctx, refinementStreamName, cfg)
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

	repo := s.newRepo()
	if err := repo.AddMessage(ctx, refinementMsg); err != nil {
		return fmt.Errorf("storing refinement response: %w", err)
	}
	s.notify(resp.IssueID)

	// On terminal agent errors (e.g. Claude 401), transition the session out
	// of 'active' so the UI stops polling and the user can start a new one.
	if resp.Type == "error" {
		if err := repo.FailSession(ctx, resp.SessionID); err != nil {
			s.logger.Warn("failing refinement session",
				zap.String("session_id", resp.SessionID.String()),
				zap.Error(err),
			)
		} else {
			s.notify(resp.IssueID)
		}
	}

	s.logger.Info("stored refinement response",
		zap.String("session_id", resp.SessionID.String()),
		zap.String("type", resp.Type),
	)
	return nil
}
