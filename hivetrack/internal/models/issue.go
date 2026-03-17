package models

import (
	"time"

	"github.com/google/uuid"
)

type IssueType string

const (
	IssueTypeEpic IssueType = "epic"
	IssueTypeTask IssueType = "task"
)

type IssueStatus string

// Software archetype statuses
const (
	IssueStatusTodo       IssueStatus = "todo"
	IssueStatusInProgress IssueStatus = "in_progress"
	IssueStatusInReview   IssueStatus = "in_review"
	IssueStatusDone       IssueStatus = "done"
	IssueStatusCancelled  IssueStatus = "cancelled"
)

// Support archetype statuses
const (
	IssueStatusOpen     IssueStatus = "open"
	IssueStatusResolved IssueStatus = "resolved"
	IssueStatusClosed   IssueStatus = "closed"
)

type IssuePriority string

const (
	IssuePriorityNone     IssuePriority = "none"
	IssuePriorityLow      IssuePriority = "low"
	IssuePriorityMedium   IssuePriority = "medium"
	IssuePriorityHigh     IssuePriority = "high"
	IssuePriorityCritical IssuePriority = "critical"
)

type IssueEstimate string

const (
	IssueEstimateNone IssueEstimate = "none"
	IssueEstimateXS   IssueEstimate = "xs"
	IssueEstimateS    IssueEstimate = "s"
	IssueEstimateM    IssueEstimate = "m"
	IssueEstimateL    IssueEstimate = "l"
	IssueEstimateXL   IssueEstimate = "xl"
)

type HoldReason string

const (
	HoldReasonWaitingOnCustomer  HoldReason = "waiting_on_customer"
	HoldReasonWaitingOnExternal  HoldReason = "waiting_on_external"
	HoldReasonBlockedByIssue     HoldReason = "blocked_by_issue"
)

type IssueVisibility string

const (
	IssueVisibilityNormal     IssueVisibility = "normal"
	IssueVisibilityRestricted IssueVisibility = "restricted"
)

type LinkType string

const (
	LinkTypeBlocks     LinkType = "blocks"
	LinkTypeDuplicates LinkType = "duplicates"
	LinkTypeRelatesTo  LinkType = "relates_to"
)

type ChecklistItem struct {
	ID   uuid.UUID `json:"id"`
	Text string    `json:"text"`
	Done bool      `json:"done"`
}

type IssueLink struct {
	ID            uuid.UUID
	SourceIssueID uuid.UUID
	TargetIssueID uuid.UUID
	LinkType      LinkType
}

type Issue struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Number    int

	Type        IssueType
	Title       string
	Description *string
	Status      IssueStatus

	OnHold     bool
	HoldReason *HoldReason
	HoldSince  *time.Time
	HoldNote   *string

	Priority IssuePriority
	Estimate IssueEstimate

	ReporterID *uuid.UUID
	ParentID   *uuid.UUID

	MilestoneID *uuid.UUID
	SprintID    *uuid.UUID

	SprintCarryCount int

	Triaged bool

	Visibility IssueVisibility

	CustomerEmail *string
	CustomerName  *string
	CustomerToken *uuid.UUID

	Checklist []ChecklistItem

	Assignees         []uuid.UUID
	Labels            []uuid.UUID
	RestrictedViewers []uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsTerminal returns true if the issue is in a terminal state.
func (i *Issue) IsTerminal() bool {
	return i.Status == IssueStatusDone ||
		i.Status == IssueStatusCancelled ||
		i.Status == IssueStatusClosed
}
