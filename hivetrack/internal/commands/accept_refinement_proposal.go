package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type AcceptRefinementProposalCommand struct {
	IssueID uuid.UUID
}

type AcceptRefinementProposalResult struct{}

func HandleAcceptRefinementProposal(ctx context.Context, cmd AcceptRefinementProposalCommand) (*AcceptRefinementProposalResult, error) {
	db := repositories.GetDbContext(ctx)

	// Load active session + messages
	session, err := db.Refinements().GetActiveSession(ctx, cmd.IssueID)
	if err != nil {
		return nil, fmt.Errorf("getting active session: %w", err)
	}
	if session == nil {
		return nil, fmt.Errorf("no active refinement session for issue %s: %w", cmd.IssueID, models.ErrNotFound)
	}

	_, messages, err := db.Refinements().GetSessionWithMessages(ctx, session.ID)
	if err != nil {
		return nil, fmt.Errorf("loading session messages: %w", err)
	}

	// Find the latest proposal
	var proposal *models.RefinementProposal
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].MessageType == models.RefinementMessageTypeProposal && messages[i].Proposal != nil {
			proposal = messages[i].Proposal
			break
		}
	}
	if proposal == nil {
		return nil, fmt.Errorf("no proposal found in session %s: %w", session.ID, models.ErrNotFound)
	}

	// Update the issue
	issue, err := db.Issues().GetByID(ctx, cmd.IssueID)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue %s: %w", cmd.IssueID, models.ErrNotFound)
	}

	issue.SetTitle(proposal.Title)
	issue.SetDescription(&proposal.Description)
	issue.SetRefined(true)
	issue.SetUpdatedAt(time.Now())
	db.Issues().Update(issue)

	// Complete the session
	if err := db.Refinements().CompleteSession(ctx, session.ID); err != nil {
		return nil, fmt.Errorf("completing session: %w", err)
	}

	// Persist issue changes
	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving changes: %w", err)
	}

	return &AcceptRefinementProposalResult{}, nil
}
