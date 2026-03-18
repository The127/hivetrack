package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type GetSprintBurndownQuery struct {
	ProjectSlug string
	SprintID    uuid.UUID
}

type BurndownDataPoint struct {
	Date      time.Time `json:"date"`
	Remaining int       `json:"remaining"`
}

type GetSprintBurndownResult struct {
	Points    []BurndownDataPoint `json:"points"`
	Total     int                 `json:"total"`
	StartDate time.Time           `json:"start_date"`
	EndDate   time.Time           `json:"end_date"`
}

var terminalStatusesByArchetype = map[models.ProjectArchetype][]string{
	models.ProjectArchetypeSoftware: {"done", "cancelled"},
	models.ProjectArchetypeSupport:  {"resolved", "closed"},
}

func HandleGetSprintBurndown(ctx context.Context, q GetSprintBurndownQuery) (*GetSprintBurndownResult, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetBySlug(ctx, q.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %q: %w", q.ProjectSlug, models.ErrNotFound)
	}

	sprint, err := db.Sprints().GetByID(ctx, q.SprintID)
	if err != nil {
		return nil, fmt.Errorf("getting sprint: %w", err)
	}
	if sprint == nil || sprint.GetProjectID() != project.GetId() {
		return nil, fmt.Errorf("sprint %s: %w", q.SprintID, models.ErrNotFound)
	}

	// Total = count of tasks in sprint
	_, total, err := db.Issues().List(ctx, repositories.NewIssueFilter().
		BySprintID(q.SprintID).
		ByType(models.IssueTypeTask))
	if err != nil {
		return nil, fmt.Errorf("counting sprint issues: %w", err)
	}

	if sprint.GetStartDate().IsZero() || sprint.GetEndDate().IsZero() {
		return &GetSprintBurndownResult{
			Points:    []BurndownDataPoint{},
			Total:     total,
			StartDate: sprint.GetStartDate(),
			EndDate:   sprint.GetEndDate(),
		}, nil
	}

	terminalStatuses := terminalStatusesByArchetype[project.GetArchetype()]

	rawPoints, err := db.IssueStatusLog().GetBurndownPoints(
		ctx, q.SprintID, sprint.GetStartDate(), sprint.GetEndDate(), terminalStatuses,
	)
	if err != nil {
		return nil, fmt.Errorf("getting burndown points: %w", err)
	}

	points := make([]BurndownDataPoint, 0, len(rawPoints))
	for _, p := range rawPoints {
		points = append(points, BurndownDataPoint{Date: p.Date, Remaining: p.Remaining})
	}

	return &GetSprintBurndownResult{
		Points:    points,
		Total:     total,
		StartDate: sprint.GetStartDate(),
		EndDate:   sprint.GetEndDate(),
	}, nil
}
