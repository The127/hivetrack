package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// StartRefinementSession starts a new refinement session for an issue.
// Returns the session ID.
func (c *Client) StartRefinementSession(ctx context.Context, slug string, number int) (string, error) {
	data, err := c.post(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/refinement/start", slug, number), nil)
	if err != nil {
		return "", err
	}
	var result struct {
		SessionID string `json:"SessionID"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("parsing result: %w", err)
	}
	return result.SessionID, nil
}

// SendRefinementMessage sends a user message in an active refinement session.
func (c *Client) SendRefinementMessage(ctx context.Context, slug string, number int, content string) error {
	_, err := c.post(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/refinement/message", slug, number), map[string]any{
		"content": content,
	})
	return err
}

// GetRefinementSession returns the current refinement session for an issue, or nil if none exists.
func (c *Client) GetRefinementSession(ctx context.Context, slug string, number int) (*RefinementSessionDetail, error) {
	data, err := c.get(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/refinement/session", slug, number), nil)
	if err != nil {
		return nil, err
	}
	if string(data) == "null" {
		return nil, nil
	}
	var session RefinementSessionDetail
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("parsing session: %w", err)
	}
	return &session, nil
}

// AcceptRefinementProposal accepts the current proposal in a refinement session.
func (c *Client) AcceptRefinementProposal(ctx context.Context, slug string, number int) error {
	_, err := c.post(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/refinement/accept", slug, number), nil)
	return err
}

// AdvanceRefinementPhase advances the refinement session to the next phase.
// Optionally pass a targetPhase to jump to a specific phase; empty string advances to next.
// Returns the new phase name.
func (c *Client) AdvanceRefinementPhase(ctx context.Context, slug string, number int, targetPhase string) (string, error) {
	var body any
	if targetPhase != "" {
		body = map[string]any{"target_phase": targetPhase}
	}
	data, err := c.post(ctx, fmt.Sprintf("/api/v1/projects/%s/issues/%d/refinement/advance-phase", slug, number), body)
	if err != nil {
		return "", err
	}
	var result struct {
		Phase string `json:"Phase"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("parsing result: %w", err)
	}
	return result.Phase, nil
}
