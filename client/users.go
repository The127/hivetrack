package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetMe returns the current authenticated user.
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	data, err := c.get(ctx, "/api/v1/users/me", nil)
	if err != nil {
		return nil, err
	}
	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("parsing user: %w", err)
	}
	return &user, nil
}

// ListUsers returns all users.
func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	data, err := c.get(ctx, "/api/v1/users", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Users []User `json:"users"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing users: %w", err)
	}
	return resp.Users, nil
}
