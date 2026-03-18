package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type UpdateSprintCommand struct {
	SprintID                 uuid.UUID
	Name                     *string
	Goal                     *string
	StartDate                *time.Time
	EndDate                  *time.Time
	Status                   *models.SprintStatus
	MoveOpenIssuesToSprintID *uuid.UUID // only used when completing; nil = backlog
}

type UpdateSprintResult struct{}

func HandleUpdateSprint(ctx context.Context, cmd UpdateSprintCommand) (*UpdateSprintResult, error) {
	db := repositories.GetDbContext(ctx)

	sprint, err := db.Sprints().GetByID(ctx, cmd.SprintID)
	if err != nil {
		return nil, fmt.Errorf("getting sprint: %w", err)
	}
	if sprint == nil {
		return nil, fmt.Errorf("sprint %s: %w", cmd.SprintID, models.ErrNotFound)
	}

	if cmd.Name != nil {
		sprint.SetName(*cmd.Name)
	}
	if cmd.Goal != nil {
		sprint.SetGoal(cmd.Goal)
	}
	if cmd.StartDate != nil {
		sprint.SetStartDate(*cmd.StartDate)
	}
	if cmd.EndDate != nil {
		sprint.SetEndDate(*cmd.EndDate)
	}

	if cmd.Status != nil {
		switch *cmd.Status {
		case models.SprintStatusActive:
			allSprints, err := db.Sprints().List(ctx, sprint.GetProjectID())
			if err != nil {
				return nil, fmt.Errorf("listing sprints: %w", err)
			}
			for _, s := range allSprints {
				if s.GetId() != sprint.GetId() && s.GetStatus() == models.SprintStatusActive {
					return nil, fmt.Errorf("another sprint is already active: %w", models.ErrConflict)
				}
			}
			sprint.SetStatus(models.SprintStatusActive)

		case models.SprintStatusCompleted:
			project, err := db.Projects().GetByID(ctx, sprint.GetProjectID())
			if err != nil {
				return nil, fmt.Errorf("getting project: %w", err)
			}
			if project == nil {
				return nil, fmt.Errorf("project not found: %w", models.ErrNotFound)
			}

			// Validate target sprint if specified.
			if cmd.MoveOpenIssuesToSprintID != nil {
				if *cmd.MoveOpenIssuesToSprintID == sprint.GetId() {
					return nil, fmt.Errorf("cannot move issues to the sprint being completed: %w", models.ErrBadRequest)
				}
				targetSprint, err := db.Sprints().GetByID(ctx, *cmd.MoveOpenIssuesToSprintID)
				if err != nil {
					return nil, fmt.Errorf("getting target sprint: %w", err)
				}
				if targetSprint == nil {
					return nil, fmt.Errorf("target sprint %s: %w", *cmd.MoveOpenIssuesToSprintID, models.ErrNotFound)
				}
			}

			filter := repositories.NewIssueFilter().BySprintID(sprint.GetId())
			issues, _, err := db.Issues().List(ctx, filter)
			if err != nil {
				return nil, fmt.Errorf("listing sprint issues: %w", err)
			}

			isTerminal := terminalChecker(project.GetArchetype())
			for _, issue := range issues {
				if !isTerminal(issue) {
					issue.SetSprintID(cmd.MoveOpenIssuesToSprintID)
					issue.SetSprintCarryCount(issue.GetSprintCarryCount() + 1)
					issue.SetUpdatedAt(time.Now())
					db.Issues().Update(issue)
				}
			}
			sprint.SetStatus(models.SprintStatusCompleted)

		default:
			sprint.SetStatus(*cmd.Status)
		}
	}

	db.Sprints().Update(sprint)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving sprint: %w", err)
	}

	return &UpdateSprintResult{}, nil
}

func terminalChecker(archetype models.ProjectArchetype) func(*models.Issue) bool {
	switch archetype {
	case models.ProjectArchetypeSupport:
		return func(i *models.Issue) bool {
			return i.GetStatus() == models.IssueStatusResolved ||
				i.GetStatus() == models.IssueStatusClosed
		}
	default: // software
		return func(i *models.Issue) bool {
			return i.GetStatus() == models.IssueStatusDone ||
				i.GetStatus() == models.IssueStatusCancelled
		}
	}
}
