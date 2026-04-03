package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type RefinementSessionStatus string

const (
	RefinementSessionActive    RefinementSessionStatus = "active"
	RefinementSessionCompleted RefinementSessionStatus = "completed"
	RefinementSessionAbandoned RefinementSessionStatus = "abandoned"
)

type RefinementMessageRole string

const (
	RefinementRoleUser      RefinementMessageRole = "user"
	RefinementRoleAssistant RefinementMessageRole = "assistant"
)

type RefinementMessageType string

const (
	RefinementMessageTypeMessage  RefinementMessageType = "message"
	RefinementMessageTypeProposal RefinementMessageType = "proposal"
)

type RefinementProposal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type RefinementSession struct {
	ID        uuid.UUID
	IssueID   uuid.UUID
	Status    RefinementSessionStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RefinementMessage struct {
	ID          uuid.UUID
	SessionID   uuid.UUID
	Role        RefinementMessageRole
	Content     string
	MessageType RefinementMessageType
	Proposal    *RefinementProposal
	CreatedAt   time.Time
}

func NewRefinementSession(issueID uuid.UUID) *RefinementSession {
	now := time.Now()
	return &RefinementSession{
		ID:        uuid.New(),
		IssueID:   issueID,
		Status:    RefinementSessionActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func NewRefinementMessage(sessionID uuid.UUID, role RefinementMessageRole, content string, messageType RefinementMessageType, proposal *RefinementProposal) *RefinementMessage {
	return &RefinementMessage{
		ID:          uuid.New(),
		SessionID:   sessionID,
		Role:        role,
		Content:     content,
		MessageType: messageType,
		Proposal:    proposal,
		CreatedAt:   time.Now(),
	}
}

// ProposalJSON returns the proposal as raw JSON bytes, or nil if no proposal.
func (m *RefinementMessage) ProposalJSON() ([]byte, error) {
	if m.Proposal == nil {
		return nil, nil
	}
	return json.Marshal(m.Proposal)
}
