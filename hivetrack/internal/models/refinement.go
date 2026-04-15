package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RefinementSessionStatus string

const (
	RefinementSessionActive    RefinementSessionStatus = "active"
	RefinementSessionCompleted RefinementSessionStatus = "completed"
	RefinementSessionAbandoned RefinementSessionStatus = "abandoned"
	RefinementSessionFailed    RefinementSessionStatus = "failed"
)

type RefinementMessageRole string

const (
	RefinementRoleUser      RefinementMessageRole = "user"
	RefinementRoleAssistant RefinementMessageRole = "assistant"
)

type RefinementMessageType string

const (
	RefinementMessageTypeMessage     RefinementMessageType = "message"
	RefinementMessageTypeProposal    RefinementMessageType = "proposal"
	RefinementMessageTypePhaseResult RefinementMessageType = "phase_result"
)

type RefinementPhase string

const (
	RefinementPhaseActorGoal          RefinementPhase = "actor_goal"
	RefinementPhaseMainScenario       RefinementPhase = "main_scenario"
	RefinementPhaseExtensions         RefinementPhase = "extensions"
	RefinementPhaseAcceptanceCriteria RefinementPhase = "acceptance_criteria"
	RefinementPhaseBddScenarios       RefinementPhase = "bdd_scenarios"
)

// RefinementPhases is the ordered list of refinement phases.
var RefinementPhases = []RefinementPhase{
	RefinementPhaseActorGoal,
	RefinementPhaseMainScenario,
	RefinementPhaseExtensions,
	RefinementPhaseAcceptanceCriteria,
	RefinementPhaseBddScenarios,
}

// NextPhase returns the phase after the given one, or an error if already at the last phase.
func NextPhase(current RefinementPhase) (RefinementPhase, error) {
	for i, p := range RefinementPhases {
		if p == current && i+1 < len(RefinementPhases) {
			return RefinementPhases[i+1], nil
		}
	}
	return "", fmt.Errorf("no next phase after %q: %w", current, ErrBadRequest)
}

// ValidPhase returns true if s is a recognized refinement phase.
func ValidPhase(s string) bool {
	for _, p := range RefinementPhases {
		if string(p) == s {
			return true
		}
	}
	return false
}

type RefinementProposal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type RefinementSession struct {
	ID           uuid.UUID
	IssueID      uuid.UUID
	Status       RefinementSessionStatus
	CurrentPhase RefinementPhase
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefinementMessage struct {
	ID          uuid.UUID
	SessionID   uuid.UUID
	Role        RefinementMessageRole
	Content     string
	MessageType RefinementMessageType
	Phase       RefinementPhase
	Proposal    *RefinementProposal
	PhaseData   map[string]interface{}
	Suggestions []string
	CreatedAt   time.Time
}

func NewRefinementSession(issueID uuid.UUID) *RefinementSession {
	now := time.Now()
	return &RefinementSession{
		ID:           uuid.New(),
		IssueID:      issueID,
		Status:       RefinementSessionActive,
		CurrentPhase: RefinementPhaseActorGoal,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func NewRefinementMessage(sessionID uuid.UUID, role RefinementMessageRole, content string, messageType RefinementMessageType, phase RefinementPhase, proposal *RefinementProposal) *RefinementMessage {
	return &RefinementMessage{
		ID:          uuid.New(),
		SessionID:   sessionID,
		Role:        role,
		Content:     content,
		MessageType: messageType,
		Phase:       phase,
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

// PhaseDataJSON returns the phase data as raw JSON bytes, or nil if none.
func (m *RefinementMessage) PhaseDataJSON() ([]byte, error) {
	if m.PhaseData == nil {
		return nil, nil
	}
	return json.Marshal(m.PhaseData)
}

// SuggestionsJSON returns suggestions as raw JSON bytes, or nil if none.
func (m *RefinementMessage) SuggestionsJSON() ([]byte, error) {
	if len(m.Suggestions) == 0 {
		return nil, nil
	}
	return json.Marshal(m.Suggestions)
}
