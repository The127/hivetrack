package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type RefinementRepository struct {
	ctx *DbContext
}

func NewRefinementRepository(ctx *DbContext) *RefinementRepository {
	return &RefinementRepository{ctx: ctx}
}

var _ repositories.RefinementRepository = (*RefinementRepository)(nil)

func (r *RefinementRepository) CreateSession(ctx context.Context, session *models.RefinementSession) error {
	return r.ctx.execDirect(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO refinement_sessions (id, issue_id, status, current_phase, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			session.ID, session.IssueID, session.Status, session.CurrentPhase, session.CreatedAt, session.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("inserting refinement session: %w", err)
		}
		return nil
	})
}

func (r *RefinementRepository) GetActiveSession(ctx context.Context, issueID uuid.UUID) (*models.RefinementSession, error) {
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, issue_id, status, current_phase, created_at, updated_at
		 FROM refinement_sessions
		 WHERE issue_id = $1 AND status = 'active'`, issueID)

	return scanRefinementSession(row)
}

func (r *RefinementRepository) GetSessionWithMessages(ctx context.Context, sessionID uuid.UUID) (*models.RefinementSession, []*models.RefinementMessage, error) {
	// Get session
	row := r.ctx.queryContext(ctx).QueryRowContext(ctx,
		`SELECT id, issue_id, status, current_phase, created_at, updated_at
		 FROM refinement_sessions WHERE id = $1`, sessionID)

	session, err := scanRefinementSession(row)
	if err != nil {
		return nil, nil, err
	}
	if session == nil {
		return nil, nil, nil
	}

	// Get messages
	rows, err := r.ctx.queryContext(ctx).QueryContext(ctx,
		`SELECT id, session_id, role, content, message_type, phase, proposal, phase_data, suggestions, created_at
		 FROM refinement_messages
		 WHERE session_id = $1
		 ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("listing refinement messages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var messages []*models.RefinementMessage
	for rows.Next() {
		msg, err := scanRefinementMessageRow(rows)
		if err != nil {
			return nil, nil, err
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterating refinement messages: %w", err)
	}

	return session, messages, nil
}

func (r *RefinementRepository) AddMessage(ctx context.Context, msg *models.RefinementMessage) error {
	proposalJSON, err := msg.ProposalJSON()
	if err != nil {
		return fmt.Errorf("marshaling proposal: %w", err)
	}

	phaseDataJSON, err := msg.PhaseDataJSON()
	if err != nil {
		return fmt.Errorf("marshaling phase data: %w", err)
	}

	suggestionsJSON, err := msg.SuggestionsJSON()
	if err != nil {
		return fmt.Errorf("marshaling suggestions: %w", err)
	}

	// For JSONB columns, nil []byte must be passed as explicit SQL NULL, not empty.
	var proposalArg any
	if proposalJSON != nil {
		proposalArg = proposalJSON
	}
	var phaseDataArg any
	if phaseDataJSON != nil {
		phaseDataArg = phaseDataJSON
	}
	var suggestionsArg any
	if suggestionsJSON != nil {
		suggestionsArg = suggestionsJSON
	}

	return r.ctx.execDirect(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO refinement_messages (id, session_id, role, content, message_type, phase, proposal, phase_data, suggestions, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			msg.ID, msg.SessionID, msg.Role, msg.Content, msg.MessageType, msg.Phase, proposalArg, phaseDataArg, suggestionsArg, msg.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("inserting refinement message: %w", err)
		}
		return nil
	})
}

func (r *RefinementRepository) CompleteSession(ctx context.Context, sessionID uuid.UUID) error {
	return r.ctx.execDirect(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx,
			`UPDATE refinement_sessions SET status = 'completed', updated_at = $1 WHERE id = $2 AND status = 'active'`,
			time.Now(), sessionID,
		)
		if err != nil {
			return fmt.Errorf("completing refinement session: %w", err)
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return fmt.Errorf("refinement session %s: %w", sessionID, models.ErrNotFound)
		}
		return nil
	})
}

func (r *RefinementRepository) UpdateSessionPhase(ctx context.Context, sessionID uuid.UUID, phase models.RefinementPhase) error {
	return r.ctx.execDirect(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx,
			`UPDATE refinement_sessions SET current_phase = $1, updated_at = $2 WHERE id = $3 AND status = 'active'`,
			phase, time.Now(), sessionID,
		)
		if err != nil {
			return fmt.Errorf("updating refinement session phase: %w", err)
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return fmt.Errorf("refinement session %s: %w", sessionID, models.ErrNotFound)
		}
		return nil
	})
}

func scanRefinementSession(row *sql.Row) (*models.RefinementSession, error) {
	var s models.RefinementSession
	err := row.Scan(&s.ID, &s.IssueID, &s.Status, &s.CurrentPhase, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning refinement session: %w", err)
	}
	return &s, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRefinementMessageRow(row rowScanner) (*models.RefinementMessage, error) {
	var msg models.RefinementMessage
	var proposalJSON []byte
	var phaseDataJSON []byte
	var suggestionsJSON []byte
	err := row.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.MessageType, &msg.Phase, &proposalJSON, &phaseDataJSON, &suggestionsJSON, &msg.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("scanning refinement message: %w", err)
	}
	if proposalJSON != nil {
		var p models.RefinementProposal
		if err := json.Unmarshal(proposalJSON, &p); err != nil {
			return nil, fmt.Errorf("unmarshaling proposal: %w", err)
		}
		msg.Proposal = &p
	}
	if phaseDataJSON != nil {
		var pd map[string]interface{}
		if err := json.Unmarshal(phaseDataJSON, &pd); err != nil {
			return nil, fmt.Errorf("unmarshaling phase data: %w", err)
		}
		msg.PhaseData = pd
	}
	if suggestionsJSON != nil {
		if err := json.Unmarshal(suggestionsJSON, &msg.Suggestions); err != nil {
			return nil, fmt.Errorf("unmarshaling suggestions: %w", err)
		}
	}
	return &msg, nil
}
