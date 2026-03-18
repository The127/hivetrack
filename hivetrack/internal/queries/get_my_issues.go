package queries

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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

	// Build projectID → slug map
	projectSlugs := map[uuid.UUID]string{}
	for _, i := range issues {
		projectSlugs[i.GetProjectID()] = ""
	}
	for pid := range projectSlugs {
		p, err := db.Projects().GetByID(ctx, pid)
		if err != nil {
			return nil, fmt.Errorf("getting project %s: %w", pid, err)
		}
		if p != nil {
			projectSlugs[pid] = p.GetSlug()
		}
	}

	var items []IssueSummary
	for _, i := range issues {
		// Exclude terminal issues
		if i.IsTerminal() {
			continue
		}
		assignees, err := resolveUsers(ctx, db, i.GetAssignees())
		if err != nil {
			return nil, fmt.Errorf("resolving assignees: %w", err)
		}
		labelInfos, err := resolveLabels(ctx, db, i.GetLabels())
		if err != nil {
			return nil, fmt.Errorf("resolving labels: %w", err)
		}
		slug := projectSlugs[i.GetProjectID()]
		items = append(items, IssueSummary{
			ID:          i.GetId(),
			Number:      i.GetNumber(),
			Type:        i.GetType(),
			Title:       i.GetTitle(),
			Status:      i.GetStatus(),
			Priority:    i.GetPriority(),
			Estimate:    i.GetEstimate(),
			Triaged:     i.GetTriaged(),
			Assignees:   assignees,
			Labels:      labelInfos,
			SprintID:    i.GetSprintID(),
			MilestoneID: i.GetMilestoneID(),
			OnHold:      i.GetOnHold(),
			ProjectSlug: &slug,
			CreatedAt:   i.GetCreatedAt(),
			UpdatedAt:   i.GetUpdatedAt(),
		})
	}

	if items == nil {
		items = []IssueSummary{}
	}

	return &GetMyIssuesResult{Items: items}, nil
}
