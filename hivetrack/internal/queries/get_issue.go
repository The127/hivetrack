package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type GetIssueQuery struct {
	ProjectSlug string
	Number      int
}

type IssueLinkInfo struct {
	ID            uuid.UUID       `json:"id"`
	SourceIssueID uuid.UUID       `json:"source_issue_id"`
	TargetIssueID uuid.UUID       `json:"target_issue_id"`
	LinkType      models.LinkType `json:"link_type"`
	// LinkedIssueNumber is the number of the other issue in the link (the one that is not the current issue).
	LinkedIssueNumber int `json:"linked_issue_number"`
}

type IssueDetail struct {
	ID             uuid.UUID              `json:"id"`
	ProjectID      uuid.UUID              `json:"project_id"`
	Number         int                    `json:"number"`
	Type           models.IssueType       `json:"type"`
	Title          string                 `json:"title"`
	Description    *string                `json:"description,omitempty"`
	Status         models.IssueStatus     `json:"status"`
	Priority       models.IssuePriority   `json:"priority"`
	Estimate       models.IssueEstimate   `json:"estimate"`
	Triaged        bool                   `json:"triaged"`
	Refined        bool                   `json:"refined"`
	Visibility     models.IssueVisibility `json:"visibility"`
	OnHold         bool                   `json:"on_hold"`
	HoldReason     *models.HoldReason     `json:"hold_reason,omitempty"`
	HoldNote       *string                `json:"hold_note,omitempty"`
	HoldSince      *time.Time             `json:"hold_since,omitempty"`
	Assignees      []UserInfo             `json:"assignees"`
	Labels         []LabelInfo            `json:"labels"`
	SprintID       *uuid.UUID             `json:"sprint_id,omitempty"`
	MilestoneID    *uuid.UUID             `json:"milestone_id,omitempty"`
	ParentID       *uuid.UUID             `json:"parent_id,omitempty"`
	ReporterID     *uuid.UUID             `json:"reporter_id,omitempty"`
	Owner          *UserInfo              `json:"owner,omitempty"`
	CancelReason   *string                `json:"cancel_reason,omitempty"`
	Checklist      []models.ChecklistItem `json:"checklist"`
	Links          []IssueLinkInfo        `json:"links"`
	ChildCount     int                    `json:"child_count"`
	ChildDoneCount int                    `json:"child_done_count"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

func HandleGetIssue(ctx context.Context, q GetIssueQuery) (*IssueDetail, error) {
	db := repositories.GetDbContext(ctx)

	project, err := db.Projects().GetBySlug(ctx, q.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project %q: %w", q.ProjectSlug, models.ErrNotFound)
	}

	issue, err := db.Issues().GetByNumber(ctx, project.GetId(), q.Number)
	if err != nil {
		return nil, fmt.Errorf("getting issue: %w", err)
	}
	if issue == nil {
		return nil, nil
	}

	assignees, err := resolveUsers(ctx, db, issue.GetAssignees())
	if err != nil {
		return nil, fmt.Errorf("resolving assignees: %w", err)
	}

	var owner *UserInfo
	if issue.GetOwnerID() != nil {
		owners, err := resolveUsers(ctx, db, []uuid.UUID{*issue.GetOwnerID()})
		if err != nil {
			return nil, fmt.Errorf("resolving owner: %w", err)
		}
		if len(owners) > 0 {
			owner = &owners[0]
		}
	}

	labels, err := resolveLabels(ctx, db, issue.GetLabels())
	if err != nil {
		return nil, fmt.Errorf("resolving labels: %w", err)
	}

	rawLinks, err := db.Issues().ListLinks(ctx, issue.GetId())
	if err != nil {
		return nil, fmt.Errorf("listing issue links: %w", err)
	}
	links := make([]IssueLinkInfo, 0, len(rawLinks))
	for _, l := range rawLinks {
		otherID := l.TargetIssueID
		if l.TargetIssueID == issue.GetId() {
			otherID = l.SourceIssueID
		}
		other, err := db.Issues().GetByID(ctx, otherID)
		if err != nil {
			return nil, fmt.Errorf("resolving linked issue: %w", err)
		}
		var linkedNumber int
		if other != nil {
			linkedNumber = other.GetNumber()
		}
		links = append(links, IssueLinkInfo{
			ID:                l.ID,
			SourceIssueID:     l.SourceIssueID,
			TargetIssueID:     l.TargetIssueID,
			LinkType:          l.LinkType,
			LinkedIssueNumber: linkedNumber,
		})
	}

	detail := &IssueDetail{
		ID:           issue.GetId(),
		ProjectID:    issue.GetProjectID(),
		Number:       issue.GetNumber(),
		Type:         issue.GetType(),
		Title:        issue.GetTitle(),
		Description:  issue.GetDescription(),
		Status:       issue.GetStatus(),
		Priority:     issue.GetPriority(),
		Estimate:     issue.GetEstimate(),
		Triaged:      issue.GetTriaged(),
		Refined:      issue.GetRefined(),
		Visibility:   issue.GetVisibility(),
		OnHold:       issue.GetOnHold(),
		HoldReason:   issue.GetHoldReason(),
		HoldNote:     issue.GetHoldNote(),
		HoldSince:    issue.GetHoldSince(),
		Assignees:    assignees,
		Labels:       labels,
		SprintID:     issue.GetSprintID(),
		MilestoneID:  issue.GetMilestoneID(),
		ParentID:     issue.GetParentID(),
		ReporterID:   issue.GetReporterID(),
		Owner:        owner,
		CancelReason: issue.GetCancelReason(),
		Checklist:    issue.GetChecklist(),
		Links:        links,
		CreatedAt:    issue.GetCreatedAt(),
		UpdatedAt:    issue.GetUpdatedAt(),
	}

	if issue.GetType() == models.IssueTypeEpic {
		children, _, err := db.Issues().List(ctx, repositories.NewIssueFilter().ByParentID(issue.GetId()))
		if err != nil {
			return nil, fmt.Errorf("listing child issues: %w", err)
		}
		detail.ChildCount = len(children)
		for _, child := range children {
			if child.IsTerminal() {
				detail.ChildDoneCount++
			}
		}
	}

	return detail, nil
}
