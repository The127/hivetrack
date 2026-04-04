package queries

import (
	"context"
	"fmt"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

// buildIssueSummary constructs an IssueSummary from a domain Issue, resolving
// assignees and labels via the DbContext. Optional fields (e.g. ProjectSlug)
// can be set by the caller after construction.
func buildIssueSummary(ctx context.Context, db repositories.DbContext, i *models.Issue) (IssueSummary, error) {
	assignees, err := resolveUsers(ctx, db, i.GetAssignees())
	if err != nil {
		return IssueSummary{}, fmt.Errorf("resolving assignees: %w", err)
	}
	labelInfos, err := resolveLabels(ctx, db, i.GetLabels())
	if err != nil {
		return IssueSummary{}, fmt.Errorf("resolving labels: %w", err)
	}
	return IssueSummary{
		ID:          i.GetId(),
		Number:      i.GetNumber(),
		Type:        i.GetType(),
		Title:       i.GetTitle(),
		Status:      i.GetStatus(),
		Priority:    i.GetPriority(),
		Estimate:    i.GetEstimate(),
		Triaged:     i.GetTriaged(),
		Refined:     i.GetRefined(),
		Assignees:   assignees,
		Labels:      labelInfos,
		SprintID:    i.GetSprintID(),
		MilestoneID: i.GetMilestoneID(),
		ParentID:    i.GetParentID(),
		Rank:        i.GetRank(),
		OnHold:      i.GetOnHold(),
		CreatedAt:   i.GetCreatedAt(),
		UpdatedAt:   i.GetUpdatedAt(),
	}, nil
}
