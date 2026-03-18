package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/authentication"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type SplitIssueCommand struct {
	IssueID   uuid.UUID
	NewTitles []string // at least 2
}

type SplitIssueResult struct {
	NewIssues []CreateIssueResult
}

func HandleSplitIssue(ctx context.Context, cmd SplitIssueCommand) (*SplitIssueResult, error) {
	db := repositories.GetDbContext(ctx)
	actor := authentication.MustGetCurrentUser(ctx)

	if len(cmd.NewTitles) < 2 {
		return nil, fmt.Errorf("at least 2 titles required to split: %w", models.ErrBadRequest)
	}

	issue, err := db.Issues().GetByID(ctx, cmd.IssueID)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue %s: %w", cmd.IssueID, models.ErrNotFound)
	}

	if issue.GetType() != models.IssueTypeTask {
		return nil, fmt.Errorf("only tasks can be split: %w", models.ErrBadRequest)
	}
	if issue.IsTerminal() {
		return nil, fmt.Errorf("cannot split a terminal issue: %w", models.ErrBadRequest)
	}

	project, err := db.Projects().GetByID(ctx, issue.GetProjectID())
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %s: %w", issue.GetProjectID(), models.ErrNotFound)
	}

	var defaultStatus models.IssueStatus
	var cancelStatus models.IssueStatus
	switch project.GetArchetype() {
	case models.ProjectArchetypeSoftware:
		defaultStatus = models.IssueStatusTodo
		cancelStatus = models.IssueStatusCancelled
	case models.ProjectArchetypeSupport:
		defaultStatus = models.IssueStatusOpen
		cancelStatus = models.IssueStatusClosed
	}

	reporterID := actor.ID

	var newIssues []*models.Issue
	for _, title := range cmd.NewTitles {
		number, err := db.Projects().NextIssueNumber(ctx, project.GetId())
		if err != nil {
			return nil, fmt.Errorf("getting next issue number: %w", err)
		}

		newIssue := models.NewIssue(
			project.GetId(), number, models.IssueTypeTask, title,
			defaultStatus, issue.GetPriority(), issue.GetEstimate(),
			&reporterID, true, models.IssueVisibilityNormal,
			nil, issue.GetSprintID(), issue.GetMilestoneID(),
			issue.GetAssignees(), issue.GetLabels(),
		)
		if issue.GetParentID() != nil {
			newIssue.SetParentID(issue.GetParentID())
		}

		db.Issues().Insert(newIssue)
		newIssues = append(newIssues, newIssue)
	}

	issue.SetStatus(cancelStatus)
	db.Issues().Update(issue)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving split: %w", err)
	}

	for _, newIssue := range newIssues {
		link := models.IssueLink{
			ID:            uuid.New(),
			SourceIssueID: issue.GetId(),
			TargetIssueID: newIssue.GetId(),
			LinkType:      models.LinkTypeRelatesTo,
		}
		if err := db.Issues().InsertLink(ctx, link); err != nil {
			return nil, fmt.Errorf("creating link to issue %s: %w", newIssue.GetId(), err)
		}
	}

	results := make([]CreateIssueResult, len(newIssues))
	for i, ni := range newIssues {
		results[i] = CreateIssueResult{
			ID:     ni.GetId(),
			Number: ni.GetNumber(),
		}
	}

	return &SplitIssueResult{NewIssues: results}, nil
}
