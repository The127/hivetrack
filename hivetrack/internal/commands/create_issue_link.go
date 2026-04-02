package commands

import (
	"context"
	"fmt"
	"time"

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

	sourceID := cmd.SourceIssueID
	targetID := cmd.TargetIssueID
	linkType := cmd.LinkType

	// Normalize is_blocked_by: swap direction and store as blocks.
	if linkType == "is_blocked_by" {
		linkType = models.LinkTypeBlocks
		sourceID, targetID = targetID, sourceID
	}

	link := models.IssueLink{
		ID:            uuid.New(),
		SourceIssueID: sourceID,
		TargetIssueID: targetID,
		LinkType:      linkType,
	}

	if err := db.Issues().InsertLink(ctx, link); err != nil {
		return nil, fmt.Errorf("inserting issue link: %w", err)
	}

	// Auto-hold: when creating a "blocks" link, set the target on hold.
	if linkType == models.LinkTypeBlocks {
		blockedIssue, err := db.Issues().GetByID(ctx, targetID)
		if err != nil {
			return nil, fmt.Errorf("getting blocked issue: %w", err)
		}
		if blockedIssue != nil && !blockedIssue.GetOnHold() {
			now := time.Now()
			reason := models.HoldReasonBlockedByIssue
			blockedIssue.SetHold(true, &reason, &now, nil)
			blockedIssue.SetUpdatedAt(now)
			db.Issues().Update(blockedIssue)
		}
	}

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving issue link: %w", err)
	}

	return &CreateIssueLinkResult{}, nil
}
