package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// ListMilestones returns all milestones in a project.
func (c *Client) ListMilestones(ctx context.Context, slug string) ([]Milestone, error) {
	data, err := c.get(ctx, "/api/v1/projects/"+slug+"/milestones", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Milestones []Milestone `json:"milestones"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing milestones: %w", err)
	}
	return resp.Milestones, nil
}

// CreateMilestoneRequest contains fields for creating a milestone.
type CreateMilestoneRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	TargetDate  string `json:"target_date,omitempty"`
}

// CreateMilestone creates a new milestone.
func (c *Client) CreateMilestone(ctx context.Context, slug string, req CreateMilestoneRequest) (string, error) {
	data, err := c.post(ctx, "/api/v1/projects/"+slug+"/milestones", req)
	if err != nil {
		return "", err
	}
	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("parsing result: %w", err)
	}
	return resp.ID, nil
}

// UpdateMilestoneRequest contains fields for updating a milestone.
type UpdateMilestoneRequest struct {
	Title       Field[string] `json:"title"`
	Description Field[string] `json:"description"`
	TargetDate  Field[string] `json:"target_date"`
	Close       Field[bool]   `json:"close"`
}

// UpdateMilestone updates an existing milestone.
func (c *Client) UpdateMilestone(ctx context.Context, slug string, milestoneID string, req UpdateMilestoneRequest) error {
	_, err := c.patchFields(ctx, fmt.Sprintf("/api/v1/projects/%s/milestones/%s", slug, milestoneID), req)
	return err
}

// DeleteMilestone deletes a milestone.
func (c *Client) DeleteMilestone(ctx context.Context, slug string, milestoneID string) error {
	_, err := c.delete(ctx, fmt.Sprintf("/api/v1/projects/%s/milestones/%s", slug, milestoneID))
	return err
}
