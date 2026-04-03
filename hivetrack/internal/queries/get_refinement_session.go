package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type GetRefinementSessionQuery struct {
	ProjectSlug string
	IssueNumber int
}

type RefinementSessionDetail struct {
	ID        uuid.UUID                      `json:"id"`
	IssueID   uuid.UUID                      `json:"issue_id"`
	Status    models.RefinementSessionStatus  `json:"status"`
	Messages  []RefinementMessageDetail       `json:"messages"`
	CreatedAt time.Time                       `json:"created_at"`
	UpdatedAt time.Time                       `json:"updated_at"`
}

type RefinementMessageDetail struct {
	ID          uuid.UUID                    `json:"id"`
	Role        models.RefinementMessageRole `json:"role"`
	Content     string                       `json:"content"`
	MessageType models.RefinementMessageType `json:"message_type"`
	Proposal    *models.RefinementProposal   `json:"proposal,omitempty"`
	CreatedAt   time.Time                    `json:"created_at"`
}

func HandleGetRefinementSession(ctx context.Context, q GetRefinementSessionQuery) (*RefinementSessionDetail, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetBySlug(ctx, q.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %q: %w", q.ProjectSlug, models.ErrNotFound)
	}

	issue, err := db.Issues().GetByNumber(ctx, project.GetId(), q.IssueNumber)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue #%d: %w", q.IssueNumber, models.ErrNotFound)
	}

	session, err := db.Refinements().GetActiveSession(ctx, issue.GetId())
	if err != nil {
		return nil, fmt.Errorf("getting active session: %w", err)
	}
	if session == nil {
		return nil, nil
	}

	_, messages, err := db.Refinements().GetSessionWithMessages(ctx, session.ID)
	if err != nil {
		return nil, fmt.Errorf("loading session messages: %w", err)
	}

	msgDetails := make([]RefinementMessageDetail, len(messages))
	for i, m := range messages {
		msgDetails[i] = RefinementMessageDetail{
			ID:          m.ID,
			Role:        m.Role,
			Content:     m.Content,
			MessageType: m.MessageType,
			Proposal:    m.Proposal,
			CreatedAt:   m.CreatedAt,
		}
	}

	return &RefinementSessionDetail{
		ID:        session.ID,
		IssueID:   session.IssueID,
		Status:    session.Status,
		Messages:  msgDetails,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}, nil
}
