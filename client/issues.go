package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ListIssuesOptions contains optional filters for listing issues.
type ListIssuesOptions struct {
	Status         string
	Priority       string
	Type           string
	Text           string
	Triaged        *bool
	Backlog        *bool
	SprintID       string
	AssigneeID     string
	LabelID        string
	ExcludeLabelID string
	Limit          int
	Offset         int
}

type listIssuesResponse struct {
	Items  []IssueSummary `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// ListIssues returns issues in a project matching the given filters.
func (c *Client) ListIssues(ctx context.Context, slug string, opts ListIssuesOptions) ([]IssueSummary, int, error) {
	q := url.Values{}
	if opts.Status != "" {
		q.Set("status", opts.Status)
	}
	if opts.Priority != "" {
		q.Set("priority", opts.Priority)
	}
	if opts.Type != "" {
		q.Set("type", opts.Type)
	}
	if opts.Text != "" {
		q.Set("text", opts.Text)
	}
	if opts.Triaged != nil {
		q.Set("triaged", fmt.Sprintf("%t", *opts.Triaged))
	}
	if opts.Backlog != nil {
		q.Set("backlog", fmt.Sprintf("%t", *opts.Backlog))
	}
	if opts.SprintID != "" {
		q.Set("sprint_id", opts.SprintID)
	}
	if opts.AssigneeID != "" {
		q.Set("assignee_id", opts.AssigneeID)
	}
	if opts.LabelID != "" {
		q.Set("label_id", opts.LabelID)
	}
	if opts.ExcludeLabelID != "" {
		q.Set("exclude_label_id", opts.ExcludeLabelID)
	}
	if opts.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	if opts.Offset > 0 {
		q.Set("offset", fmt.Sprintf("%d", opts.Offset))
	}

	data, err := c.get(ctx, "/api/v1/projects/"+slug+"/issues", q)
	if err != nil {
		return nil, 0, err
	}
	var resp listIssuesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, 0, fmt.Errorf("parsing issues: %w", err)
	}
	return resp.Items, resp.Total, nil
}

// GetIssue returns full details of a single issue.
func (c *Client) GetIssue(ctx context.Context, slug string, number int) (*IssueDetail, error) {
	data, err := c.get(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d", slug, number), nil)
	if err != nil {
		return nil, err
	}
	var issue IssueDetail
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("parsing issue: %w", err)
	}
	return &issue, nil
}

// GetMyIssues returns all issues assigned to the current user.
func (c *Client) GetMyIssues(ctx context.Context) ([]IssueSummary, error) {
	data, err := c.get(ctx, "/api/v1/me/issues", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Items []IssueSummary `json:"items"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing issues: %w", err)
	}
	return resp.Items, nil
}

// CreateIssueRequest contains the fields for creating an issue.
type CreateIssueRequest struct {
	Title       string   `json:"title"`
	Type        string   `json:"type,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
	Estimate    string   `json:"estimate,omitempty"`
	SprintID    string   `json:"sprint_id,omitempty"`
	MilestoneID string   `json:"milestone_id,omitempty"`
	ParentID    string   `json:"parent_id,omitempty"`
	AssigneeIDs []string `json:"assignee_ids,omitempty"`
	LabelIDs    []string `json:"label_ids,omitempty"`
}

// CreateIssueResult contains the ID and number of the created issue.
type CreateIssueResult struct {
	ID     string `json:"ID"`
	Number int    `json:"Number"`
}

// CreateIssue creates a new issue in a project.
func (c *Client) CreateIssue(ctx context.Context, slug string, req CreateIssueRequest) (*CreateIssueResult, error) {
	data, err := c.post(ctx, "/api/v1/projects/"+slug+"/issues", req)
	if err != nil {
		return nil, err
	}
	var result CreateIssueResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing result: %w", err)
	}
	return &result, nil
}

// UpdateIssueRequest contains fields for updating an issue.
// Use Set() for values, Null() to clear nullable fields, leave absent to skip.
type UpdateIssueRequest struct {
	Title       Field[string]   `json:"title"`
	Description Field[string]   `json:"description"`
	Status      Field[string]   `json:"status"`
	Priority    Field[string]   `json:"priority"`
	Estimate    Field[string]   `json:"estimate"`
	SprintID    Field[string]   `json:"sprint_id"`
	MilestoneID Field[string]   `json:"milestone_id"`
	ParentID    Field[string]   `json:"parent_id"`
	AssigneeIDs Field[[]string] `json:"assignee_ids"`
	LabelIDs    Field[[]string] `json:"label_ids"`
	OwnerID     Field[string]   `json:"owner_id"`
}

// UpdateIssue updates an existing issue.
func (c *Client) UpdateIssue(ctx context.Context, slug string, number int, req UpdateIssueRequest) error {
	_, err := c.patchFields(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d", slug, number), req)
	return err
}

// DeleteIssue permanently deletes an issue.
func (c *Client) DeleteIssue(ctx context.Context, slug string, number int) error {
	_, err := c.delete(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d", slug, number))
	return err
}

// TriageIssueRequest contains fields for triaging an issue.
type TriageIssueRequest struct {
	Status      string  `json:"status"`
	SprintID    *string `json:"sprint_id,omitempty"`
	MilestoneID *string `json:"milestone_id,omitempty"`
	Priority    *string `json:"priority,omitempty"`
	Estimate    *string `json:"estimate,omitempty"`
}

// TriageIssue triages an untriaged issue.
func (c *Client) TriageIssue(ctx context.Context, slug string, number int, req TriageIssueRequest) error {
	_, err := c.post(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/triage", slug, number), req)
	return err
}

// RefineIssue marks an issue as refined.
func (c *Client) RefineIssue(ctx context.Context, slug string, number int) error {
	_, err := c.post(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/refine", slug, number), nil)
	return err
}

// SplitIssueResult contains the new issues created from a split.
type SplitIssueResult struct {
	NewIssues []CreateIssueResult `json:"new_issues"`
}

// SplitIssue splits an issue into multiple new issues.
func (c *Client) SplitIssue(ctx context.Context, slug string, number int, titles []string) (*SplitIssueResult, error) {
	data, err := c.post(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/split", slug, number), map[string]any{
		"titles": titles,
	})
	if err != nil {
		return nil, err
	}
	var result SplitIssueResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing result: %w", err)
	}
	return &result, nil
}

// AddIssueLink adds a link between two issues.
func (c *Client) AddIssueLink(ctx context.Context, slug string, number int, linkType LinkType, targetNumber int) error {
	_, err := c.post(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/links", slug, number), map[string]any{
		"link_type":     linkType,
		"target_number": targetNumber,
	})
	return err
}

// AddChecklistItem adds a checklist item to a task.
func (c *Client) AddChecklistItem(ctx context.Context, slug string, number int, text string) (string, error) {
	data, err := c.post(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/checklist", slug, number), map[string]any{
		"text": text,
	})
	if err != nil {
		return "", err
	}
	var resp struct {
		ID string `json:"ID"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("parsing result: %w", err)
	}
	return resp.ID, nil
}

// UpdateChecklistItemRequest contains fields for updating a checklist item.
type UpdateChecklistItemRequest struct {
	Text *string `json:"text,omitempty"`
	Done *bool   `json:"done,omitempty"`
}

// UpdateChecklistItem updates a checklist item.
func (c *Client) UpdateChecklistItem(ctx context.Context, slug string, number int, itemID string, req UpdateChecklistItemRequest) error {
	_, err := c.patch(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/checklist/%s", slug, number, itemID), req)
	return err
}

// RemoveChecklistItem removes a checklist item.
func (c *Client) RemoveChecklistItem(ctx context.Context, slug string, number int, itemID string) error {
	_, err := c.delete(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/checklist/%s", slug, number, itemID))
	return err
}

// BatchUpdateIssuesRequest contains fields for batch updating issues.
type BatchUpdateIssuesRequest struct {
	Numbers       []int            `json:"numbers"`
	Status        Field[string]    `json:"status"`
	Priority      Field[string]    `json:"priority"`
	Estimate      Field[string]    `json:"estimate"`
	SprintID      Field[string]    `json:"sprint_id"`
	ClearSprintID Field[bool]      `json:"clear_sprint_id"`
	MilestoneID   Field[string]    `json:"milestone_id"`
	AssigneeIDs   Field[[]string]  `json:"assignee_ids"`
	LabelIDs      Field[[]string]  `json:"label_ids"`
}

// BatchUpdateIssuesResult contains the count of updated issues.
type BatchUpdateIssuesResult struct {
	Updated int `json:"Updated"`
}

// BatchUpdateIssues applies the same changes to multiple issues atomically.
func (c *Client) BatchUpdateIssues(ctx context.Context, slug string, req BatchUpdateIssuesRequest) (*BatchUpdateIssuesResult, error) {
	data, err := c.postFields(ctx, "/api/v1/projects/"+slug+"/issues/batch-update", req)
	if err != nil {
		return nil, err
	}
	var result BatchUpdateIssuesResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing result: %w", err)
	}
	return &result, nil
}
