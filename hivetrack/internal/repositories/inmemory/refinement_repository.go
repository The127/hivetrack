package inmemory

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
)

type RefinementRepository struct {
	sessions map[uuid.UUID]*models.RefinementSession
	messages map[uuid.UUID]*models.RefinementMessage
}

func NewRefinementRepository() *RefinementRepository {
	return &RefinementRepository{
		sessions: make(map[uuid.UUID]*models.RefinementSession),
		messages: make(map[uuid.UUID]*models.RefinementMessage),
	}
}

func (r *RefinementRepository) CreateSession(_ context.Context, session *models.RefinementSession) error {
	r.sessions[session.ID] = session
	return nil
}

func (r *RefinementRepository) GetActiveSession(_ context.Context, issueID uuid.UUID) (*models.RefinementSession, error) {
	for _, s := range r.sessions {
		if s.IssueID == issueID && s.Status == models.RefinementSessionActive {
			return s, nil
		}
	}
	return nil, nil
}

func (r *RefinementRepository) GetSessionWithMessages(_ context.Context, sessionID uuid.UUID) (*models.RefinementSession, []*models.RefinementMessage, error) {
	session, ok := r.sessions[sessionID]
	if !ok {
		return nil, nil, nil
	}

	var msgs []*models.RefinementMessage
	for _, m := range r.messages {
		if m.SessionID == sessionID {
			msgs = append(msgs, m)
		}
	}
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].CreatedAt.Before(msgs[j].CreatedAt)
	})

	return session, msgs, nil
}

func (r *RefinementRepository) AddMessage(_ context.Context, msg *models.RefinementMessage) error {
	r.messages[msg.ID] = msg
	return nil
}

func (r *RefinementRepository) CompleteSession(_ context.Context, sessionID uuid.UUID) error {
	session, ok := r.sessions[sessionID]
	if !ok {
		return models.ErrNotFound
	}
	session.Status = models.RefinementSessionCompleted
	session.UpdatedAt = time.Now()
	return nil
}
