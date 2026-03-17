package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type DeleteIssueCommand struct {
	IssueID uuid.UUID
}

type DeleteIssueResult struct{}

func HandleDeleteIssue(ctx context.Context, cmd DeleteIssueCommand) (*DeleteIssueResult, error) {
	db := repositories.GetDbContext(ctx)

	issue, err := db.Issues().GetByID(ctx, cmd.IssueID)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue %s: %w", cmd.IssueID, models.ErrNotFound)
	}

	if err := db.Issues().Delete(ctx, cmd.IssueID); err != nil {
		return nil, fmt.Errorf("deleting issue: %w", err)
	}

	if err := db.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing: %w", err)
	}

	return &DeleteIssueResult{}, nil
}
