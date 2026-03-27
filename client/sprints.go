package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// ListSprints returns all sprints in a project.
func (c *Client) ListSprints(ctx context.Context, slug string) ([]Sprint, error) {
	data, err := c.get(ctx, "/api/v1/projects/"+slug+"/sprints", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Sprints []Sprint `json:"sprints"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing sprints: %w", err)
	}
	return resp.Sprints, nil
}

// CreateSprintRequest contains fields for creating a sprint.
type CreateSprintRequest struct {
	Name      string `json:"name"`
	Goal      string `json:"goal,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

// CreateSprint creates a new sprint.
func (c *Client) CreateSprint(ctx context.Context, slug string, req CreateSprintRequest) (string, error) {
	data, err := c.post(ctx, "/api/v1/projects/"+slug+"/sprints", req)
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

// UpdateSprintRequest contains fields for updating a sprint.
type UpdateSprintRequest struct {
	Name                     *string
	Goal                     *string
	StartDate                *string
	EndDate                  *string
	Status                   *string
	Force                    bool
	MoveOpenIssuesToSprintID *string
}

func (r UpdateSprintRequest) toMap() map[string]any {
	m := map[string]any{}
	if r.Name != nil {
		m["name"] = *r.Name
	}
	if r.Goal != nil {
		m["goal"] = *r.Goal
	}
	if r.StartDate != nil {
		m["start_date"] = *r.StartDate
	}
	if r.EndDate != nil {
		m["end_date"] = *r.EndDate
	}
	if r.Status != nil {
		m["status"] = *r.Status
	}
	if r.Force {
		m["force"] = true
	}
	if r.MoveOpenIssuesToSprintID != nil {
		m["move_open_issues_to_sprint_id"] = *r.MoveOpenIssuesToSprintID
	}
	return m
}

// UpdateSprint updates an existing sprint.
func (c *Client) UpdateSprint(ctx context.Context, slug string, sprintID string, req UpdateSprintRequest) error {
	_, err := c.patch(ctx, fmt.Sprintf("/api/v1/projects/%s/sprints/%s", slug, sprintID), req.toMap())
	return err
}

// DeleteSprint permanently deletes a sprint.
func (c *Client) DeleteSprint(ctx context.Context, slug string, sprintID string) error {
	_, err := c.delete(ctx, fmt.Sprintf("/api/v1/projects/%s/sprints/%s", slug, sprintID))
	return err
}

// GetSprintBurndown returns burndown chart data for a sprint.
func (c *Client) GetSprintBurndown(ctx context.Context, slug string, sprintID string) (*BurndownData, error) {
	data, err := c.get(ctx, fmt.Sprintf("/api/v1/projects/%s/sprints/%s/burndown", slug, sprintID), nil)
	if err != nil {
		return nil, err
	}
	var burndown BurndownData
	if err := json.Unmarshal(data, &burndown); err != nil {
		return nil, fmt.Errorf("parsing burndown: %w", err)
	}
	return &burndown, nil
}
