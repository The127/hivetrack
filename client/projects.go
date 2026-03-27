package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// ListProjects returns all projects the current user has access to.
func (c *Client) ListProjects(ctx context.Context) ([]ProjectSummary, error) {
	data, err := c.get(ctx, "/api/v1/projects", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Items []ProjectSummary `json:"items"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing projects: %w", err)
	}
	return resp.Items, nil
}

// GetProject returns details of a single project.
func (c *Client) GetProject(ctx context.Context, slug string) (*Project, error) {
	data, err := c.get(ctx, "/api/v1/projects/"+slug, nil)
	if err != nil {
		return nil, err
	}
	var project Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("parsing project: %w", err)
	}
	return &project, nil
}

// CreateProjectRequest contains the fields for creating a project.
type CreateProjectRequest struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Archetype   string `json:"archetype"`
	Description string `json:"description,omitempty"`
}

// CreateProject creates a new project.
func (c *Client) CreateProject(ctx context.Context, req CreateProjectRequest) (string, error) {
	data, err := c.post(ctx, "/api/v1/projects", req)
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

// UpdateProjectRequest contains fields for updating a project.
// WIP limits: use Set(-1) to clear, Set(N) to set, leave absent to skip.
type UpdateProjectRequest struct {
	Name               Field[string] `json:"name"`
	Description        Field[string] `json:"description"`
	Archived           Field[bool]   `json:"archived"`
	WipLimitInProgress Field[int]    `json:"wip_limit_in_progress"`
	WipLimitInReview   Field[int]    `json:"wip_limit_in_review"`
}

// UpdateProject updates a project by its ID (UUID).
func (c *Client) UpdateProject(ctx context.Context, projectID string, req UpdateProjectRequest) error {
	_, err := c.patchFields(ctx, "/api/v1/projects/"+projectID, req)
	return err
}

// DeleteProject deletes a project by its ID (UUID).
func (c *Client) DeleteProject(ctx context.Context, projectID string) error {
	_, err := c.delete(ctx, "/api/v1/projects/"+projectID)
	return err
}

// AddProjectMember adds a user to a project with a role.
func (c *Client) AddProjectMember(ctx context.Context, slug string, userID string, role ProjectRole) error {
	_, err := c.post(ctx, "/api/v1/projects/"+slug+"/members", map[string]any{
		"user_id": userID,
		"role":    role,
	})
	return err
}

// RemoveProjectMember removes a user from a project.
func (c *Client) RemoveProjectMember(ctx context.Context, slug string, userID string) error {
	_, err := c.delete(ctx, fmt.Sprintf("/api/v1/projects/%s/members/%s", slug, userID))
	return err
}
