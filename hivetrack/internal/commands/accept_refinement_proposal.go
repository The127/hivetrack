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

// HandleAcceptRefinementProposal returns a handler that depends on a RefinementPublisher.
// notify is invoked post-commit so real-time subscribers can refetch the session.
func HandleAcceptRefinementProposal(publisher RefinementPublisher, notify func(uuid.UUID)) func(context.Context, AcceptRefinementProposalCommand) (*AcceptRefinementProposalResult, error) {
	return func(ctx context.Context, cmd AcceptRefinementProposalCommand) (*AcceptRefinementProposalResult, error) {
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
		notify(cmd.IssueID)

		// Signal Hivemind to clean up its session (best-effort, don't fail the accept)
		_ = publisher.PublishRefinementAccept(ctx, session.ID)

		// Load project for slug
		project, err := db.Projects().GetByID(ctx, issue.GetProjectID())
		if err != nil {
			return nil, fmt.Errorf("getting project: %w", err)
		}

		// Publish story.refined event for downstream pipeline (best-effort)
		event := StoryRefinedEvent{
			StoryID:     cmd.IssueID.String(),
			ProjectID:   issue.GetProjectID().String(),
			ProjectSlug: project.GetSlug(),
			IssueNumber: issue.GetNumber(),
			Title:       proposal.Title,
		}
		// Extract structured data from the last message's phase_data if available
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].PhaseData != nil {
				if v, ok := messages[i].PhaseData["actor"].(string); ok {
					event.Actor = v
				}
				if v, ok := messages[i].PhaseData["goal"].(string); ok {
					event.Goal = v
				}
				if v, ok := messages[i].PhaseData["main_success_scenario"].([]interface{}); ok {
					for _, s := range v {
						if str, ok := s.(string); ok {
							event.MainSuccessScenario = append(event.MainSuccessScenario, str)
						}
					}
				}
				if v, ok := messages[i].PhaseData["preconditions"].([]interface{}); ok {
					for _, s := range v {
						if str, ok := s.(string); ok {
							event.Preconditions = append(event.Preconditions, str)
						}
					}
				}
				if v, ok := messages[i].PhaseData["acceptance_criteria"].([]interface{}); ok {
					for _, s := range v {
						if str, ok := s.(string); ok {
							event.AcceptanceCriteria = append(event.AcceptanceCriteria, str)
						}
					}
				}
				if v, ok := messages[i].PhaseData["extensions"].([]interface{}); ok {
					for _, s := range v {
						if str, ok := s.(string); ok {
							event.Extensions = append(event.Extensions, str)
						}
					}
				}
				// The final proposal's phase_data should have everything
				break
			}
		}
		_ = publisher.PublishStoryRefined(ctx, event)

		return &AcceptRefinementProposalResult{}, nil
	}
}
