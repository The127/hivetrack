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
	Items []IssueSummary `json:"items"`
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
		summary, err := buildIssueSummary(ctx, db, i)
		if err != nil {
			return nil, err
		}
		slug := projectSlugs[i.GetProjectID()]
		summary.ProjectSlug = &slug
		items = append(items, summary)
	}

	if items == nil {
		items = []IssueSummary{}
	}

	return &GetMyIssuesResult{Items: items}, nil
}
