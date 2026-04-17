package client

import (
	"time"
)

// Enums

type IssueType string

const (
	IssueTypeEpic IssueType = "epic"
	IssueTypeTask IssueType = "task"
)

type IssueStatus string

const (
	IssueStatusTodo       IssueStatus = "todo"
	IssueStatusInProgress IssueStatus = "in_progress"
	IssueStatusInReview   IssueStatus = "in_review"
	IssueStatusDone       IssueStatus = "done"
	IssueStatusCancelled  IssueStatus = "cancelled"
	IssueStatusOpen       IssueStatus = "open"
	IssueStatusResolved   IssueStatus = "resolved"
	IssueStatusClosed     IssueStatus = "closed"
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

type IssueVisibility string

const (
	IssueVisibilityNormal     IssueVisibility = "normal"
	IssueVisibilityRestricted IssueVisibility = "restricted"
)

type LinkType string

const (
	LinkTypeBlocks      LinkType = "blocks"
	LinkTypeIsBlockedBy LinkType = "is_blocked_by"
	LinkTypeDuplicates  LinkType = "duplicates"
	LinkTypeRelatesTo   LinkType = "relates_to"
)

type HoldReason string

const (
	HoldReasonWaitingOnCustomer HoldReason = "waiting_on_customer"
	HoldReasonWaitingOnExternal HoldReason = "waiting_on_external"
	HoldReasonBlockedByIssue    HoldReason = "blocked_by_issue"
)

type ProjectArchetype string

const (
	ProjectArchetypeSoftware ProjectArchetype = "software"
	ProjectArchetypeSupport  ProjectArchetype = "support"
)

type ProjectRole string

const (
	ProjectRoleAdmin  ProjectRole = "project_admin"
	ProjectRoleMember ProjectRole = "project_member"
	ProjectRoleViewer ProjectRole = "viewer"
)

type SprintStatus string

const (
	SprintStatusPlanning  SprintStatus = "planning"
	SprintStatusActive    SprintStatus = "active"
	SprintStatusCompleted SprintStatus = "completed"
)

// Shared sub-types

type UserInfo struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type LabelInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type ChecklistItem struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

type IssueLinkInfo struct {
	ID                string   `json:"id"`
	SourceIssueID     string   `json:"source_issue_id"`
	TargetIssueID     string   `json:"target_issue_id"`
	LinkType          LinkType `json:"link_type"`
	LinkedIssueNumber int      `json:"linked_issue_number"`
}

// Response types

type Project struct {
	ID        string           `json:"id"`
	Slug      string           `json:"slug"`
	Name      string           `json:"name"`
	Archetype ProjectArchetype `json:"archetype"`
	Members   []ProjectMember  `json:"members"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type ProjectMember struct {
	UserID      string      `json:"user_id"`
	Email       string      `json:"email"`
	DisplayName string      `json:"display_name"`
	Role        ProjectRole `json:"role"`
}

type ProjectSummary struct {
	Slug      string           `json:"slug"`
	Name      string           `json:"name"`
	Archetype ProjectArchetype `json:"archetype"`
}

type IssueDetail struct {
	ID             string          `json:"id"`
	ProjectID      string          `json:"project_id"`
	Number         int             `json:"number"`
	Type           IssueType       `json:"type"`
	Title          string          `json:"title"`
	Description    *string         `json:"description,omitempty"`
	Status         IssueStatus     `json:"status"`
	Priority       IssuePriority   `json:"priority"`
	Estimate       IssueEstimate   `json:"estimate"`
	Triaged        bool            `json:"triaged"`
	Refined        bool            `json:"refined"`
	Visibility     IssueVisibility `json:"visibility"`
	OnHold         bool            `json:"on_hold"`
	HoldReason     *HoldReason     `json:"hold_reason,omitempty"`
	HoldNote       *string         `json:"hold_note,omitempty"`
	HoldSince      *time.Time      `json:"hold_since,omitempty"`
	Assignees      []UserInfo      `json:"assignees"`
	Labels         []LabelInfo     `json:"labels"`
	SprintID       *string         `json:"sprint_id,omitempty"`
	MilestoneID    *string         `json:"milestone_id,omitempty"`
	ParentID       *string         `json:"parent_id,omitempty"`
	ReporterID     *string         `json:"reporter_id,omitempty"`
	Owner          *UserInfo       `json:"owner,omitempty"`
	CancelReason   *string         `json:"cancel_reason,omitempty"`
	Checklist      []ChecklistItem `json:"checklist"`
	Links          []IssueLinkInfo `json:"links"`
	ChildCount     int             `json:"child_count"`
	ChildDoneCount int             `json:"child_done_count"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type IssueSummary struct {
	ID          string        `json:"id"`
	Number      int           `json:"number"`
	Type        IssueType     `json:"type"`
	Title       string        `json:"title"`
	Status      IssueStatus   `json:"status"`
	Priority    IssuePriority `json:"priority"`
	Estimate    IssueEstimate `json:"estimate"`
	Triaged     bool          `json:"triaged"`
	Refined     bool          `json:"refined"`
	Assignees   []UserInfo    `json:"assignees"`
	Labels      []LabelInfo   `json:"labels"`
	SprintID    *string       `json:"sprint_id,omitempty"`
	MilestoneID *string       `json:"milestone_id,omitempty"`
	ParentID    *string       `json:"parent_id,omitempty"`
	Rank        *string       `json:"rank,omitempty"`
	OnHold      bool          `json:"on_hold"`
	ProjectSlug *string       `json:"project_slug,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type Sprint struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Goal      string       `json:"goal"`
	Status    SprintStatus `json:"status"`
	StartDate string       `json:"start_date"`
	EndDate   string       `json:"end_date"`
}

type Milestone struct {
	ID               string  `json:"id"`
	Title            string  `json:"title"`
	Description      *string `json:"description,omitempty"`
	TargetDate       *string `json:"target_date,omitempty"`
	ClosedAt         *string `json:"closed_at,omitempty"`
	IssueCount       int     `json:"issue_count"`
	ClosedIssueCount int     `json:"closed_issue_count"`
}

type Comment struct {
	ID          string `json:"id"`
	AuthorName  string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
	Body        string `json:"body"`
	CreatedAt   string `json:"created_at"`
}

type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	IsAdmin     bool   `json:"is_admin"`
}

type BurndownPoint struct {
	Date      string `json:"date"`
	Remaining int    `json:"remaining"`
}

type BurndownData struct {
	Total          int             `json:"total"`
	StartRemaining int             `json:"start_remaining"`
	EndRemaining   int             `json:"end_remaining"`
	Points         []BurndownPoint `json:"points"`
}

// Refinement types

type RefinementSessionStatus string

const (
	RefinementSessionActive    RefinementSessionStatus = "active"
	RefinementSessionCompleted RefinementSessionStatus = "completed"
	RefinementSessionAbandoned RefinementSessionStatus = "abandoned"
	RefinementSessionFailed    RefinementSessionStatus = "failed"
)

type RefinementPhase string

const (
	RefinementPhaseActorGoal          RefinementPhase = "actor_goal"
	RefinementPhaseMainScenario       RefinementPhase = "main_scenario"
	RefinementPhaseExtensions         RefinementPhase = "extensions"
	RefinementPhaseAcceptanceCriteria RefinementPhase = "acceptance_criteria"
	RefinementPhaseBddScenarios       RefinementPhase = "bdd_scenarios"
)

type RefinementMessageRole string

const (
	RefinementRoleUser      RefinementMessageRole = "user"
	RefinementRoleAssistant RefinementMessageRole = "assistant"
)

type RefinementMessageType string

const (
	RefinementMessageTypeMessage     RefinementMessageType = "message"
	RefinementMessageTypeProposal    RefinementMessageType = "proposal"
	RefinementMessageTypePhaseResult RefinementMessageType = "phase_result"
)

type RefinementProposal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type RefinementMessageDetail struct {
	ID          string                `json:"id"`
	Role        RefinementMessageRole `json:"role"`
	Content     string                `json:"content"`
	MessageType RefinementMessageType `json:"message_type"`
	Phase       RefinementPhase       `json:"phase"`
	Proposal    *RefinementProposal   `json:"proposal,omitempty"`
	PhaseData   map[string]any        `json:"phase_data,omitempty"`
	Suggestions []string              `json:"suggestions,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
}

type RefinementSessionDetail struct {
	ID              string                    `json:"id"`
	IssueID         string                    `json:"issue_id"`
	Status          RefinementSessionStatus   `json:"status"`
	CurrentPhase    RefinementPhase           `json:"current_phase"`
	Messages        []RefinementMessageDetail `json:"messages"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
	PartialResponse string                    `json:"partial_response"`
	IsGenerating    bool                      `json:"is_generating"`
}

// Drone / Hivemind types

type HivemindConfig struct {
	GrpcURL string `json:"grpc_url"`
}

type Drone struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	ProjectSlug    string   `json:"project_slug"`
	Status         string   `json:"status"`
	Capabilities   []string `json:"capabilities"`
	MaxConcurrency int      `json:"max_concurrency"`
	RegisteredAt   string   `json:"registered_at"`
	LastSeenAt     *string  `json:"last_seen_at,omitempty"`
}

type CreateDroneTokenRequest struct {
	Capabilities   []string `json:"capabilities"`
	MaxConcurrency int      `json:"max_concurrency,omitempty"`
}

type CreateDroneTokenResult struct {
	Token string `json:"token"`
}
