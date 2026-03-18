package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type GetMilestonesQuery struct {
	ProjectSlug string
}

type MilestoneSummary struct {
	ID               uuid.UUID  `json:"id"`
	Title            string     `json:"title"`
	Description      *string    `json:"description"`
	TargetDate       *time.Time `json:"target_date"`
	ClosedAt         *time.Time `json:"closed_at"`
	IssueCount       int        `json:"issue_count"`
	ClosedIssueCount int        `json:"closed_issue_count"`
}

type GetMilestonesResult struct {
	Milestones []MilestoneSummary `json:"milestones"`
}

func HandleGetMilestones(ctx context.Context, q GetMilestonesQuery) (*GetMilestonesResult, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetBySlug(ctx, q.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %q: %w", q.ProjectSlug, models.ErrNotFound)
	}

	milestones, err := db.Milestones().List(ctx, project.GetId())
	if err != nil {
		return nil, fmt.Errorf("listing milestones: %w", err)
	}

	progress, err := db.Milestones().CountByMilestone(ctx, project.GetId())
	if err != nil {
		return nil, fmt.Errorf("counting milestone issues: %w", err)
	}

	summaries := make([]MilestoneSummary, 0, len(milestones))
	for _, m := range milestones {
		p := progress[m.GetId()]
		summaries = append(summaries, MilestoneSummary{
			ID:               m.GetId(),
			Title:            m.GetTitle(),
			Description:      m.GetDescription(),
			TargetDate:       m.GetTargetDate(),
			ClosedAt:         m.GetClosedAt(),
			IssueCount:       p.IssueCount,
			ClosedIssueCount: p.ClosedIssueCount,
		})
	}

	return &GetMilestonesResult{Milestones: summaries}, nil
}
