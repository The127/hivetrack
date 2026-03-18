package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/authentication"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type CreateIssueCommand struct {
	ProjectSlug string
	Title       string
	Type        models.IssueType
	Priority    *models.IssuePriority
	Estimate    *models.IssueEstimate
	Status      *models.IssueStatus
	Description *string
	AssigneeIDs []uuid.UUID
	LabelIDs    []uuid.UUID
	SprintID    *uuid.UUID
	MilestoneID *uuid.UUID
	ParentID    *uuid.UUID
}

type CreateIssueResult struct {
	ID     uuid.UUID
	Number int
}

func HandleCreateIssue(ctx context.Context, cmd CreateIssueCommand) (*CreateIssueResult, error) {
	db := repositories.GetDbContext(ctx)
	actor := authentication.MustGetCurrentUser(ctx)

	project, err := db.Projects().GetBySlug(ctx, cmd.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %q: %w", cmd.ProjectSlug, models.ErrNotFound)
	}

	number, err := db.Projects().NextIssueNumber(ctx, project.GetId())
	if err != nil {
		return nil, fmt.Errorf("getting next issue number: %w", err)
	}

	// Determine default status
	status := models.IssueStatus("")
	if cmd.Status != nil {
		status = *cmd.Status
	} else {
		switch project.GetArchetype() {
		case models.ProjectArchetypeSoftware:
			status = models.IssueStatusTodo
		case models.ProjectArchetypeSupport:
			status = models.IssueStatusOpen
		}
	}

	// Quick-capture: triaged=false only if no placement context given
	triaged := cmd.Status != nil || cmd.SprintID != nil || cmd.MilestoneID != nil || cmd.ParentID != nil

	priority := models.IssuePriorityNone
	if cmd.Priority != nil {
		priority = *cmd.Priority
	}
	estimate := models.IssueEstimateNone
	if cmd.Estimate != nil {
		estimate = *cmd.Estimate
	}

	reporterID := actor.ID

	if cmd.ParentID != nil {
		if cmd.Type != models.IssueTypeTask {
			return nil, fmt.Errorf("only tasks can have a parent: %w", models.ErrBadRequest)
		}
		parent, err := db.Issues().GetByID(ctx, *cmd.ParentID)
		if err != nil {
			return nil, fmt.Errorf("getting parent issue: %w", err)
		}
		if parent == nil {
			return nil, fmt.Errorf("parent issue %s: %w", cmd.ParentID, models.ErrNotFound)
		}
		if parent.GetType() != models.IssueTypeEpic {
			return nil, fmt.Errorf("parent must be an epic: %w", models.ErrBadRequest)
		}
		if parent.GetProjectID() != project.GetId() {
			return nil, fmt.Errorf("parent must be in the same project: %w", models.ErrBadRequest)
		}
	}

	issue := models.NewIssue(
		project.GetId(), number, cmd.Type, cmd.Title,
		status, priority, estimate,
		&reporterID, triaged, models.IssueVisibilityNormal,
		cmd.Description, cmd.SprintID, cmd.MilestoneID,
		cmd.AssigneeIDs, cmd.LabelIDs,
	)

	if cmd.ParentID != nil {
		issue.SetParentID(cmd.ParentID)
	}

	db.Issues().Insert(issue)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving issue: %w", err)
	}

	return &CreateIssueResult{
		ID:     issue.GetId(),
		Number: issue.GetNumber(),
	}, nil
}
