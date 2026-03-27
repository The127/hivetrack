package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ListLabels returns all labels in a project.
func (c *Client) ListLabels(ctx context.Context, slug string) ([]LabelInfo, error) {
	data, err := c.get(ctx, "/api/v1/projects/"+slug+"/labels", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Labels []LabelInfo `json:"labels"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing labels: %w", err)
	}
	return resp.Labels, nil
}

// CreateLabel creates a new label in a project.
func (c *Client) CreateLabel(ctx context.Context, slug string, name, color string) (string, error) {
	data, err := c.post(ctx, "/api/v1/projects/"+slug+"/labels", map[string]any{
		"name":  name,
		"color": color,
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

// UpdateLabelRequest contains fields for updating a label.
type UpdateLabelRequest struct {
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
}

// UpdateLabel updates an existing label.
func (c *Client) UpdateLabel(ctx context.Context, slug string, labelID string, req UpdateLabelRequest) error {
	_, err := c.patch(ctx, fmt.Sprintf("/api/v1/projects/%s/labels/%s", slug, labelID), req)
	return err
}

// DeleteLabel deletes a label.
func (c *Client) DeleteLabel(ctx context.Context, slug string, labelID string) error {
	_, err := c.delete(ctx, fmt.Sprintf("/api/v1/projects/%s/labels/%s", slug, labelID))
	return err
}

// ResolveLabelNames resolves comma-separated label names to their UUIDs.
// Returns an error if any label name is not found.
func (c *Client) ResolveLabelNames(ctx context.Context, slug string, names string) ([]string, error) {
	labels, err := c.ListLabels(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("fetching labels: %w", err)
	}

	nameToID := make(map[string]string, len(labels))
	for _, l := range labels {
		nameToID[strings.ToLower(l.Name)] = l.ID
	}

	var ids []string
	for _, name := range strings.Split(names, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		id, found := nameToID[strings.ToLower(name)]
		if !found {
			return nil, fmt.Errorf("label %q not found in project %s", name, slug)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
