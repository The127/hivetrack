package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/The127/mediatr"
	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/authentication"
	"github.com/the127/hivetrack/internal/events"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type UpdateIssueCommand struct {
	IssueID       uuid.UUID
	Title         *string
	Description   *string
	Status        *models.IssueStatus
	Priority      *models.IssuePriority
	Estimate      *models.IssueEstimate
	AssigneeIDs   []uuid.UUID
	LabelIDs      []uuid.UUID
	SprintID      *uuid.UUID
	ClearSprintID bool
	MilestoneID   *uuid.UUID
	ParentID      *uuid.UUID
	ClearParentID bool
	Hold          HoldUpdate
	Visibility    *models.IssueVisibility
	Rank          *string
	OwnerID       *uuid.UUID
	ClearOwnerID  bool
	CancelReason  *string
	Refined       *bool
}

type UpdateIssueResult struct{}

func HandleUpdateIssue(ctx context.Context, cmd UpdateIssueCommand) (*UpdateIssueResult, error) {
	db := repositories.GetDbContext(ctx)
	actor, _ := authentication.GetCurrentUser(ctx)

	issue, err := db.Issues().GetByID(ctx, cmd.IssueID)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue %s: %w", cmd.IssueID, models.ErrNotFound)
	}

	oldStatus := issue.GetStatus()

	if err := applyFieldUpdates(ctx, db, issue, cmd, actor); err != nil {
		return nil, err
	}

	if err := handleStatusChangeSideEffects(ctx, db, issue, cmd, oldStatus, actor); err != nil {
		return nil, err
	}

	issue.SetUpdatedAt(time.Now())
	db.Issues().Update(issue)

	if err := db.SaveChanges(ctx); err != nil {
		return nil, fmt.Errorf("saving issue: %w", err)
	}

	if err := recordStatusTransition(ctx, db, issue, cmd, oldStatus); err != nil {
		return nil, err
	}

	return &UpdateIssueResult{}, nil
}

// issueFieldPatch holds the fields shared between single and batch issue updates.
type issueFieldPatch struct {
	Status        *models.IssueStatus
	Priority      *models.IssuePriority
	Estimate      *models.IssueEstimate
	AssigneeIDs   []uuid.UUID
	LabelIDs      []uuid.UUID
	SprintID      *uuid.UUID
	ClearSprintID bool
	MilestoneID   *uuid.UUID
	Hold          HoldUpdate
}

// applyCommonFieldPatch applies the shared field-patching logic used by both
// single and batch issue updates.
func applyCommonFieldPatch(issue *models.Issue, patch issueFieldPatch) {
	if patch.Status != nil {
		issue.SetStatus(*patch.Status)
	}
	if patch.Priority != nil {
		issue.SetPriority(*patch.Priority)
	}
	if patch.Estimate != nil {
		issue.SetEstimate(*patch.Estimate)
	}
	if patch.AssigneeIDs != nil {
		issue.SetAssignees(patch.AssigneeIDs)
	}
	if patch.LabelIDs != nil {
		issue.SetLabels(patch.LabelIDs)
	}
	if patch.ClearSprintID {
		issue.SetSprintID(nil)
	} else if patch.SprintID != nil {
		issue.SetSprintID(patch.SprintID)
	}
	if patch.MilestoneID != nil {
		issue.SetMilestoneID(patch.MilestoneID)
	}
	if patch.Hold.OnHold != nil {
		if *patch.Hold.OnHold {
			now := time.Now()
			issue.SetHold(true, patch.Hold.HoldReason, &now, patch.Hold.HoldNote)
		} else {
			issue.SetHold(false, nil, nil, nil)
		}
	}
}

// applyFieldUpdates patches all fields on the issue from the command.
func applyFieldUpdates(ctx context.Context, db repositories.DbContext, issue *models.Issue, cmd UpdateIssueCommand, actor authentication.CurrentUser) error {
	if cmd.Title != nil {
		issue.SetTitle(*cmd.Title)
	}
	if cmd.Description != nil {
		issue.SetDescription(cmd.Description)
	}

	applyCommonFieldPatch(issue, issueFieldPatch{
		Status:        cmd.Status,
		Priority:      cmd.Priority,
		Estimate:      cmd.Estimate,
		AssigneeIDs:   cmd.AssigneeIDs,
		LabelIDs:      cmd.LabelIDs,
		SprintID:      cmd.SprintID,
		ClearSprintID: cmd.ClearSprintID,
		MilestoneID:   cmd.MilestoneID,
		Hold:          cmd.Hold,
	})

	if err := applyParentUpdate(ctx, db, issue, cmd); err != nil {
		return err
	}
	if cmd.Visibility != nil {
		issue.SetVisibility(*cmd.Visibility)
	}
	if cmd.Rank != nil {
		issue.SetRank(cmd.Rank)
	}
	if cmd.ClearOwnerID {
		issue.SetOwnerID(nil)
	} else if cmd.OwnerID != nil {
		issue.SetOwnerID(cmd.OwnerID)
	}
	if cmd.CancelReason != nil {
		issue.SetCancelReason(cmd.CancelReason)
	}
	if err := applyRefinedUpdate(ctx, db, issue, cmd, actor); err != nil {
		return err
	}
	return nil
}

// applyParentUpdate validates and sets the parent issue reference.
func applyParentUpdate(ctx context.Context, db repositories.DbContext, issue *models.Issue, cmd UpdateIssueCommand) error {
	if cmd.ClearParentID {
		issue.SetParentID(nil)
		return nil
	}
	if cmd.ParentID == nil {
		return nil
	}
	if issue.GetType() != models.IssueTypeTask {
		return fmt.Errorf("only tasks can have a parent: %w", models.ErrBadRequest)
	}
	parent, err := db.Issues().GetByID(ctx, *cmd.ParentID)
	if err != nil {
		return fmt.Errorf("getting parent issue: %w", err)
	}
	if parent == nil {
		return fmt.Errorf("parent issue %s: %w", cmd.ParentID, models.ErrNotFound)
	}
	if parent.GetType() != models.IssueTypeEpic {
		return fmt.Errorf("parent must be an epic: %w", models.ErrBadRequest)
	}
	if parent.GetProjectID() != issue.GetProjectID() {
		return fmt.Errorf("parent must be in the same project: %w", models.ErrBadRequest)
	}
	issue.SetParentID(cmd.ParentID)
	return nil
}

// applyRefinedUpdate validates permissions and sets the refined flag.
func applyRefinedUpdate(ctx context.Context, db repositories.DbContext, issue *models.Issue, cmd UpdateIssueCommand, actor authentication.CurrentUser) error {
	if cmd.Refined == nil {
		return nil
	}
	if issue.GetType() == models.IssueTypeEpic {
		return models.NewDomainError("refined_not_supported_for_epics", models.ErrBadRequest)
	}
	if !actor.IsAdmin {
		isViewer, err := actorIsViewerOnProject(ctx, db, issue.GetProjectID(), actor.ID)
		if err != nil {
			return err
		}
		if isViewer {
			return fmt.Errorf("actor lacks write permission to mark issue as refined: %w", models.ErrForbidden)
		}
	}
	if *cmd.Refined && issue.GetRefined() {
		return models.NewDomainError("already_refined", models.ErrConflict)
	}
	issue.SetRefined(*cmd.Refined)
	return nil
}

// handleStatusChangeSideEffects runs auto-assign, auto-triage, and auto-clear-hold
// logic triggered by a status change.
func handleStatusChangeSideEffects(ctx context.Context, db repositories.DbContext, issue *models.Issue, cmd UpdateIssueCommand, oldStatus models.IssueStatus, actor authentication.CurrentUser) error {
	if cmd.Status == nil {
		return nil
	}
	newStatus := *cmd.Status

	if oldStatus == models.IssueStatusTodo && newStatus == models.IssueStatusInProgress {
		if err := dispatchAutoAssignEvent(ctx, issue, oldStatus, newStatus, actor); err != nil {
			return err
		}
	}

	if newStatus != oldStatus && !issue.GetTriaged() {
		issue.SetTriaged(true)
	}

	if models.IsTerminalStatus(newStatus) {
		if err := autoClearBlockedHolds(ctx, db, issue); err != nil {
			return err
		}
	}

	return nil
}

// dispatchAutoAssignEvent fires the status-changed event so the auto-assign
// handler can add the actor as an assignee.
func dispatchAutoAssignEvent(ctx context.Context, issue *models.Issue, oldStatus, newStatus models.IssueStatus, actor authentication.CurrentUser) error {
	m, ok := getMediatorFromContext(ctx)
	if !ok {
		return nil
	}
	if err := mediatr.SendEvent(ctx, m, events.IssueStatusChangedEvent{
		Issue:     issue,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		ActorID:   actor.ID,
	}); err != nil {
		return fmt.Errorf("auto-assign on status change: %w", err)
	}
	return nil
}

// recordStatusTransition writes to the status log for burndown tracking.
func recordStatusTransition(ctx context.Context, db repositories.DbContext, issue *models.Issue, cmd UpdateIssueCommand, oldStatus models.IssueStatus) error {
	if cmd.Status == nil || *cmd.Status == oldStatus {
		return nil
	}
	if err := db.IssueStatusLog().Insert(ctx, issue.GetId(), string(*cmd.Status), time.Now()); err != nil {
		return fmt.Errorf("logging issue status: %w", err)
	}
	return nil
}

func actorIsViewerOnProject(ctx context.Context, db repositories.DbContext, projectID, actorID uuid.UUID) (bool, error) {
	member, err := db.Projects().GetMember(ctx, projectID, actorID)
	if err != nil {
		return false, fmt.Errorf("getting project member: %w", err)
	}
	return member != nil && member.Role == models.ProjectRoleViewer, nil
}

// autoClearBlockedHolds clears on_hold on issues blocked by this one,
// but only if all other blockers are also in a terminal state.
func autoClearBlockedHolds(ctx context.Context, db repositories.DbContext, issue *models.Issue) error {
	links, err := db.Issues().ListLinks(ctx, issue.GetId())
	if err != nil {
		return fmt.Errorf("listing links for auto-clear: %w", err)
	}
	for _, link := range links {
		if link.LinkType != models.LinkTypeBlocks || link.SourceIssueID != issue.GetId() {
			continue
		}
		blockedIssue, err := db.Issues().GetByID(ctx, link.TargetIssueID)
		if err != nil {
			return fmt.Errorf("getting blocked issue: %w", err)
		}
		if blockedIssue == nil || !blockedIssue.GetOnHold() {
			continue
		}
		reason := blockedIssue.GetHoldReason()
		if reason != nil && *reason != models.HoldReasonBlockedByIssue {
			continue
		}
		hasActiveBlockers, err := hasOtherActiveBlockers(ctx, db, blockedIssue.GetId(), issue.GetId())
		if err != nil {
			return err
		}
		if !hasActiveBlockers {
			blockedIssue.SetHold(false, nil, nil, nil)
			blockedIssue.SetUpdatedAt(time.Now())
			db.Issues().Update(blockedIssue)
		}
	}
	return nil
}

func hasOtherActiveBlockers(ctx context.Context, db repositories.DbContext, blockedID, resolvedBlockerID uuid.UUID) (bool, error) {
	links, err := db.Issues().ListLinks(ctx, blockedID)
	if err != nil {
		return false, fmt.Errorf("listing blocked issue links: %w", err)
	}
	for _, link := range links {
		if link.LinkType != models.LinkTypeBlocks || link.TargetIssueID != blockedID || link.SourceIssueID == resolvedBlockerID {
			continue
		}
		blocker, err := db.Issues().GetByID(ctx, link.SourceIssueID)
		if err != nil {
			return false, fmt.Errorf("getting blocker: %w", err)
		}
		if blocker != nil && !models.IsTerminalStatus(blocker.GetStatus()) {
			return true, nil
		}
	}
	return false, nil
}
