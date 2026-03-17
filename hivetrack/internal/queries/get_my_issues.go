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
			ID:          i.ID,
			Number:      i.Number,
			Type:        i.Type,
			Title:       i.Title,
			Status:      i.Status,
			Priority:    i.Priority,
			Estimate:    i.Estimate,
			Triaged:     i.Triaged,
			Assignees:   i.Assignees,
			Labels:      i.Labels,
			SprintID:    i.SprintID,
			MilestoneID: i.MilestoneID,
			OnHold:      i.OnHold,
			CreatedAt:   i.CreatedAt,
			UpdatedAt:   i.UpdatedAt,
		})
	}

	if items == nil {
		items = []IssueSummary{}
	}

	return &GetMyIssuesResult{Items: items}, nil
}
