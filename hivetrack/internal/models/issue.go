package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/change"
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
	HoldReasonWaitingOnCustomer HoldReason = "waiting_on_customer"
	HoldReasonWaitingOnExternal HoldReason = "waiting_on_external"
	HoldReasonBlockedByIssue    HoldReason = "blocked_by_issue"
)

type IssueVisibility string

const (
	IssueVisibilityNormal     IssueVisibility = "normal"
	IssueVisibilityRestricted IssueVisibility = "restricted"
)

type LinkType string

const (
	LinkTypeBlocks         LinkType = "blocks"
	LinkTypeIsBlockedBy    LinkType = "is_blocked_by"
	LinkTypeDuplicates     LinkType = "duplicates"
	LinkTypeIsDuplicatedBy LinkType = "is_duplicated_by"
	LinkTypeRelatesTo      LinkType = "relates_to"
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

type IssueChange int

const (
	IssueChangeTitle             IssueChange = iota
	IssueChangeDescription       IssueChange = iota
	IssueChangeStatus            IssueChange = iota
	IssueChangeHold              IssueChange = iota // covers onHold + holdReason + holdSince + holdNote
	IssueChangePriority          IssueChange = iota
	IssueChangeEstimate          IssueChange = iota
	IssueChangeMilestoneID       IssueChange = iota
	IssueChangeSprintID          IssueChange = iota
	IssueChangeSprintCarryCount  IssueChange = iota
	IssueChangeTriaged           IssueChange = iota
	IssueChangeRefined           IssueChange = iota
	IssueChangeVisibility        IssueChange = iota
	IssueChangeChecklist         IssueChange = iota
	IssueChangeAssignees         IssueChange = iota
	IssueChangeLabels            IssueChange = iota
	IssueChangeRestrictedViewers IssueChange = iota
	IssueChangeRank              IssueChange = iota
	IssueChangeParentID          IssueChange = iota
	IssueChangeOwnerID           IssueChange = iota
	IssueChangeCancelReason      IssueChange = iota
)

type Issue struct {
	BaseModel
	change.List[IssueChange]

	projectID uuid.UUID
	number    int

	issueType   IssueType
	title       string
	description *string
	status      IssueStatus

	onHold     bool
	holdReason *HoldReason
	holdSince  *time.Time
	holdNote   *string

	priority IssuePriority
	estimate IssueEstimate

	reporterID  *uuid.UUID
	ownerID     *uuid.UUID
	parentID    *uuid.UUID
	milestoneID *uuid.UUID
	sprintID    *uuid.UUID

	sprintCarryCount int

	triaged bool
	refined bool

	visibility IssueVisibility

	customerEmail *string
	customerName  *string
	customerToken *uuid.UUID

	rank         *string
	cancelReason *string

	checklist         []ChecklistItem
	assignees         []uuid.UUID
	labels            []uuid.UUID
	restrictedViewers []uuid.UUID
}

func NewIssue(projectID uuid.UUID, number int, issueType IssueType, title string,
	status IssueStatus, priority IssuePriority, estimate IssueEstimate,
	reporterID *uuid.UUID, triaged bool, visibility IssueVisibility,
	description *string, sprintID, milestoneID *uuid.UUID,
	assignees, labels []uuid.UUID) *Issue {
	return &Issue{
		BaseModel:   NewBaseModel(),
		List:        change.NewList[IssueChange](),
		projectID:   projectID,
		number:      number,
		issueType:   issueType,
		title:       title,
		description: description,
		status:      status,
		priority:    priority,
		estimate:    estimate,
		reporterID:  reporterID,
		ownerID:     reporterID,
		sprintID:    sprintID,
		milestoneID: milestoneID,
		triaged:     triaged,
		refined:     false,
		visibility:  visibility,
		assignees:   assignees,
		labels:      labels,
		checklist:   []ChecklistItem{},
	}
}

func NewIssueFromDB(
	id uuid.UUID, createdAt, updatedAt time.Time, version any,
	projectID uuid.UUID, number int,
	issueType IssueType, title string, description *string, status IssueStatus,
	onHold bool, holdReason *HoldReason, holdSince *time.Time, holdNote *string,
	priority IssuePriority, estimate IssueEstimate,
	reporterID, ownerID, parentID, milestoneID, sprintID *uuid.UUID,
	sprintCarryCount int, triaged bool, refined bool, visibility IssueVisibility,
	customerEmail, customerName *string, customerToken *uuid.UUID,
	rank *string, cancelReason *string,
	checklist []ChecklistItem, assignees, labels, restrictedViewers []uuid.UUID,
) *Issue {
	return &Issue{
		BaseModel:         NewBaseModelFromDB(id, createdAt, updatedAt, version),
		List:              change.NewList[IssueChange](),
		projectID:         projectID,
		number:            number,
		issueType:         issueType,
		title:             title,
		description:       description,
		status:            status,
		onHold:            onHold,
		holdReason:        holdReason,
		holdSince:         holdSince,
		holdNote:          holdNote,
		priority:          priority,
		estimate:          estimate,
		reporterID:        reporterID,
		ownerID:           ownerID,
		parentID:          parentID,
		milestoneID:       milestoneID,
		sprintID:          sprintID,
		sprintCarryCount:  sprintCarryCount,
		triaged:           triaged,
		refined:           refined,
		visibility:        visibility,
		customerEmail:     customerEmail,
		customerName:      customerName,
		customerToken:     customerToken,
		rank:              rank,
		cancelReason:      cancelReason,
		checklist:         checklist,
		assignees:         assignees,
		labels:            labels,
		restrictedViewers: restrictedViewers,
	}
}

// Getters
func (i *Issue) GetProjectID() uuid.UUID           { return i.projectID }
func (i *Issue) GetNumber() int                    { return i.number }
func (i *Issue) GetType() IssueType                { return i.issueType }
func (i *Issue) GetTitle() string                  { return i.title }
func (i *Issue) GetDescription() *string           { return i.description }
func (i *Issue) GetStatus() IssueStatus            { return i.status }
func (i *Issue) GetOnHold() bool                   { return i.onHold }
func (i *Issue) GetHoldReason() *HoldReason        { return i.holdReason }
func (i *Issue) GetHoldSince() *time.Time          { return i.holdSince }
func (i *Issue) GetHoldNote() *string              { return i.holdNote }
func (i *Issue) GetPriority() IssuePriority        { return i.priority }
func (i *Issue) GetEstimate() IssueEstimate        { return i.estimate }
func (i *Issue) GetReporterID() *uuid.UUID         { return i.reporterID }
func (i *Issue) GetOwnerID() *uuid.UUID            { return i.ownerID }
func (i *Issue) GetParentID() *uuid.UUID           { return i.parentID }
func (i *Issue) GetMilestoneID() *uuid.UUID        { return i.milestoneID }
func (i *Issue) GetSprintID() *uuid.UUID           { return i.sprintID }
func (i *Issue) GetSprintCarryCount() int          { return i.sprintCarryCount }
func (i *Issue) GetTriaged() bool                  { return i.triaged }
func (i *Issue) GetRefined() bool                  { return i.refined }
func (i *Issue) GetVisibility() IssueVisibility    { return i.visibility }
func (i *Issue) GetCustomerEmail() *string         { return i.customerEmail }
func (i *Issue) GetCustomerName() *string          { return i.customerName }
func (i *Issue) GetCustomerToken() *uuid.UUID      { return i.customerToken }
func (i *Issue) GetChecklist() []ChecklistItem     { return i.checklist }
func (i *Issue) GetAssignees() []uuid.UUID         { return i.assignees }
func (i *Issue) GetLabels() []uuid.UUID            { return i.labels }
func (i *Issue) GetRank() *string                  { return i.rank }
func (i *Issue) GetCancelReason() *string          { return i.cancelReason }
func (i *Issue) GetRestrictedViewers() []uuid.UUID { return i.restrictedViewers }

// Setters
func (i *Issue) SetTitle(v string) {
	if i.title == v {
		return
	}
	i.title = v
	i.TrackChange(IssueChangeTitle)
}

func (i *Issue) SetDescription(v *string) {
	i.description = v
	i.TrackChange(IssueChangeDescription)
}

func (i *Issue) SetStatus(v IssueStatus) {
	if i.status == v {
		return
	}
	i.status = v
	i.TrackChange(IssueChangeStatus)
}

// SetHold sets all hold-related fields together.
func (i *Issue) SetHold(onHold bool, reason *HoldReason, since *time.Time, note *string) {
	i.onHold = onHold
	i.holdReason = reason
	i.holdSince = since
	i.holdNote = note
	i.TrackChange(IssueChangeHold)
}

func (i *Issue) SetPriority(v IssuePriority) {
	if i.priority == v {
		return
	}
	i.priority = v
	i.TrackChange(IssueChangePriority)
}

func (i *Issue) SetEstimate(v IssueEstimate) {
	if i.estimate == v {
		return
	}
	i.estimate = v
	i.TrackChange(IssueChangeEstimate)
}

func (i *Issue) SetMilestoneID(v *uuid.UUID) {
	i.milestoneID = v
	i.TrackChange(IssueChangeMilestoneID)
}

func (i *Issue) SetSprintID(v *uuid.UUID) {
	i.sprintID = v
	i.TrackChange(IssueChangeSprintID)
}

func (i *Issue) SetSprintCarryCount(v int) {
	i.sprintCarryCount = v
	i.TrackChange(IssueChangeSprintCarryCount)
}

func (i *Issue) SetTriaged(v bool) {
	if i.triaged == v {
		return
	}
	i.triaged = v
	i.TrackChange(IssueChangeTriaged)
}

func (i *Issue) SetRefined(v bool) {
	if i.refined == v {
		return
	}
	i.refined = v
	i.TrackChange(IssueChangeRefined)
}

func (i *Issue) SetVisibility(v IssueVisibility) {
	if i.visibility == v {
		return
	}
	i.visibility = v
	i.TrackChange(IssueChangeVisibility)
}

func (i *Issue) SetChecklist(v []ChecklistItem) {
	i.checklist = v
	i.TrackChange(IssueChangeChecklist)
}

func (i *Issue) SetAssignees(v []uuid.UUID) {
	i.assignees = v
	i.TrackChange(IssueChangeAssignees)
}

func (i *Issue) SetLabels(v []uuid.UUID) {
	i.labels = v
	i.TrackChange(IssueChangeLabels)
}

func (i *Issue) SetRestrictedViewers(v []uuid.UUID) {
	i.restrictedViewers = v
	i.TrackChange(IssueChangeRestrictedViewers)
}

func (i *Issue) SetRank(v *string) {
	i.rank = v
	i.TrackChange(IssueChangeRank)
}

func (i *Issue) SetParentID(v *uuid.UUID) {
	i.parentID = v
	i.TrackChange(IssueChangeParentID)
}

func (i *Issue) SetOwnerID(v *uuid.UUID) {
	i.ownerID = v
	i.TrackChange(IssueChangeOwnerID)
}

func (i *Issue) SetCancelReason(v *string) {
	i.cancelReason = v
	i.TrackChange(IssueChangeCancelReason)
}

// IsTerminal returns true if the issue is in a terminal state.
func (i *Issue) IsTerminal() bool {
	return i.status == IssueStatusDone ||
		i.status == IssueStatusCancelled ||
		i.status == IssueStatusClosed
}
