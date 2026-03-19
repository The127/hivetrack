package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/authentication"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type RefineIssueCommand struct {
	IssueID uuid.UUID
}

type RefineIssueResult struct{}

func HandleRefineIssue(ctx context.Context, cmd RefineIssueCommand) (*RefineIssueResult, error) {
	db := repositories.GetDbContext(ctx)

	issue, err := db.Issues().GetByID(ctx, cmd.IssueID)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue %s: %w", cmd.IssueID, models.ErrNotFound)
	}

	issue.SetRefined(true)
	issue.SetUpdatedAt(time.Now())

	db.Issues().Update(issue)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving issue: %w", err)
	}

	actor := authentication.MustGetCurrentUser(ctx)
	entry := models.NewAuditLogEntry(issue.GetId(), "refined", actor.ID)
	if err := db.AuditLog().Insert(ctx, entry); err != nil {
		return nil, fmt.Errorf("inserting audit log: %w", err)
	}

	return &RefineIssueResult{}, nil
}
