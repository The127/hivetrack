package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// ListComments returns comments on an issue.
func (c *Client) ListComments(ctx context.Context, slug string, number int) ([]Comment, int, error) {
	data, err := c.get(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/comments", slug, number), nil)
	if err != nil {
		return nil, 0, err
	}
	var resp struct {
		Items []Comment `json:"items"`
		Total int       `json:"total"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, 0, fmt.Errorf("parsing comments: %w", err)
	}
	return resp.Items, resp.Total, nil
}

// CreateComment adds a comment to an issue.
func (c *Client) CreateComment(ctx context.Context, slug string, number int, body string) error {
	_, err := c.post(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/comments", slug, number), map[string]any{
		"body": body,
	})
	return err
}

// UpdateComment updates a comment's body.
func (c *Client) UpdateComment(ctx context.Context, slug string, number int, commentID string, body string) error {
	_, err := c.patch(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/comments/%s", slug, number, commentID), map[string]any{
		"body": body,
	})
	return err
}

// DeleteComment deletes a comment.
func (c *Client) DeleteComment(ctx context.Context, slug string, number int, commentID string) error {
	_, err := c.delete(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/comments/%s", slug, number, commentID))
	return err
}
