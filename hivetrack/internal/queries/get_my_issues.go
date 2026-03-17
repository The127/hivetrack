package queries

import (
	"context"
	"fmt"

	"github.com/the127/hivetrack/internal/authentication"
	"github.com/the127/hivetrack/internal/repositories"
)

type GetMyIssuesQuery struct{}

type GetMyIssuesResult struct {
	Items []IssueSummary
}

func HandleGetMyIssues(ctx context.Context, _ GetMyIssuesQuery) (*GetMyIssuesResult, error) {
	db := repositories.GetDbContext(ctx)
	actor := authentication.MustGetCurrentUser(ctx)

	filter := repositories.NewIssueFilter().ByAssigneeID(actor.ID)

	issues, _, err := db.Issues().List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing issues: %w", err)
	}

	var items []IssueSummary
	for _, i := range issues {
		// Exclude terminal issues
		if i.IsTerminal() {
			continue
		}
		items = append(items, IssueSummary{
			ID:          i.GetId(),
			Number:      i.GetNumber(),
			Type:        i.GetType(),
			Title:       i.GetTitle(),
			Status:      i.GetStatus(),
			Priority:    i.GetPriority(),
			Estimate:    i.GetEstimate(),
			Triaged:     i.GetTriaged(),
			Assignees:   i.GetAssignees(),
			Labels:      i.GetLabels(),
			SprintID:    i.GetSprintID(),
			MilestoneID: i.GetMilestoneID(),
			OnHold:      i.GetOnHold(),
			CreatedAt:   i.GetCreatedAt(),
			UpdatedAt:   i.GetUpdatedAt(),
		})
	}

	if items == nil {
		items = []IssueSummary{}
	}

	return &GetMyIssuesResult{Items: items}, nil
}
