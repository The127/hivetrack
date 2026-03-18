package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type CreateIssueLinkCommand struct {
	SourceIssueID uuid.UUID
	TargetIssueID uuid.UUID
	LinkType      models.LinkType
}

type CreateIssueLinkResult struct{}

func HandleCreateIssueLink(ctx context.Context, cmd CreateIssueLinkCommand) (*CreateIssueLinkResult, error) {
	db := repositories.GetDbContext(ctx)

	link := models.IssueLink{
		ID:            uuid.New(),
		SourceIssueID: cmd.SourceIssueID,
		TargetIssueID: cmd.TargetIssueID,
		LinkType:      cmd.LinkType,
	}

	if err := db.Issues().InsertLink(ctx, link); err != nil {
		return nil, fmt.Errorf("inserting issue link: %w", err)
	}

	return &CreateIssueLinkResult{}, nil
}
