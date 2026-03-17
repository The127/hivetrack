package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type TriageIssueCommand struct {
	IssueID     uuid.UUID
	Status      models.IssueStatus
	SprintID    *uuid.UUID
	MilestoneID *uuid.UUID
}

type TriageIssueResult struct{}

func HandleTriageIssue(ctx context.Context, cmd TriageIssueCommand) (*TriageIssueResult, error) {
	db := repositories.GetDbContext(ctx)

	issue, err := db.Issues().GetByID(ctx, cmd.IssueID)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue %s: %w", cmd.IssueID, models.ErrNotFound)
	}

	issue.Triaged = true
	issue.Status = cmd.Status
	if cmd.SprintID != nil {
		issue.SprintID = cmd.SprintID
	}
	if cmd.MilestoneID != nil {
		issue.MilestoneID = cmd.MilestoneID
	}
	issue.UpdatedAt = time.Now()

	if err := db.Issues().Update(ctx, issue); err != nil {
		return nil, fmt.Errorf("updating issue: %w", err)
	}

	if err := db.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing: %w", err)
	}

	return &TriageIssueResult{}, nil
}
