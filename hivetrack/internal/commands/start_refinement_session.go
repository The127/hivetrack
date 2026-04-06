package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

// RefinementPublisher publishes refinement requests to the messaging infrastructure.
type RefinementPublisher interface {
	PublishRefinementRequest(ctx context.Context, req RefinementPublishRequest) error
	PublishRefinementAccept(ctx context.Context, sessionID uuid.UUID) error
	PublishStoryRefined(ctx context.Context, event StoryRefinedEvent) error
}

// StoryRefinedEvent is published when a user accepts a refinement proposal.
type StoryRefinedEvent struct {
	StoryID             string   `json:"story_id"`
	ProjectID           string   `json:"project_id"`
	ProjectSlug         string   `json:"project_slug"`
	IssueNumber         int      `json:"issue_number"`
	Title               string   `json:"title"`
	Actor               string   `json:"actor"`
	Goal                string   `json:"goal"`
	MainSuccessScenario []string `json:"main_success_scenario"`
	Preconditions       []string `json:"preconditions"`
	AcceptanceCriteria  []string `json:"acceptance_criteria"`
	Extensions          []string `json:"extensions"`
}

// RefinementPublishRequest is the data sent to Hivemind for refinement.
type RefinementPublishRequest struct {
	SessionID   uuid.UUID
	IssueID     uuid.UUID
	ProjectSlug string
	Title       string
	Description *string
	Phase       string
	Messages    []RefinementChatMessage
}

// RefinementChatMessage is a single message in the refinement conversation.
type RefinementChatMessage struct {
	Role    string
	Content string
}

type StartRefinementSessionCommand struct {
	IssueID uuid.UUID
}

type StartRefinementSessionResult struct {
	SessionID uuid.UUID
}

// NewStartRefinementSessionHandler returns a handler that depends on a RefinementPublisher.
func NewStartRefinementSessionHandler(publisher RefinementPublisher) func(context.Context, StartRefinementSessionCommand) (*StartRefinementSessionResult, error) {
	return func(ctx context.Context, cmd StartRefinementSessionCommand) (*StartRefinementSessionResult, error) {
		db := repositories.GetDbContext(ctx)

		// Check no active session exists
		existing, err := db.Refinements().GetActiveSession(ctx, cmd.IssueID)
		if err != nil {
			return nil, fmt.Errorf("checking active session: %w", err)
		}
		if existing != nil {
			return nil, fmt.Errorf("issue %s already has an active refinement session: %w", cmd.IssueID, models.ErrConflict)
		}

		// Load the issue
		issue, err := db.Issues().GetByID(ctx, cmd.IssueID)
		if err != nil {
			return nil, fmt.Errorf("getting issue: %w", err)
		}
		if issue == nil {
			return nil, fmt.Errorf("issue %s: %w", cmd.IssueID, models.ErrNotFound)
		}

		// Load the project for slug
		project, err := db.Projects().GetByID(ctx, issue.GetProjectID())
		if err != nil {
			return nil, fmt.Errorf("getting project: %w", err)
		}

		// Create session
		session := models.NewRefinementSession(cmd.IssueID)
		if err := db.Refinements().CreateSession(ctx, session); err != nil {
			return nil, fmt.Errorf("creating refinement session: %w", err)
		}

		// Publish initial request to NATS
		if err := publisher.PublishRefinementRequest(ctx, RefinementPublishRequest{
			SessionID:   session.ID,
			IssueID:     cmd.IssueID,
			ProjectSlug: project.GetSlug(),
			Title:       issue.GetTitle(),
			Description: issue.GetDescription(),
			Phase:       string(models.RefinementPhaseActorGoal),
			Messages:    nil,
		}); err != nil {
			return nil, fmt.Errorf("publishing refinement request: %w", err)
		}

		return &StartRefinementSessionResult{SessionID: session.ID}, nil
	}
}
