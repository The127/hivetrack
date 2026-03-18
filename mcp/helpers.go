package mcp

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// errMissing returns a descriptive error for missing required arguments.
func errMissing(fields string) error {
	return fmt.Errorf("missing required argument(s): %s", fields)
}

// intArg extracts a number argument as int (JSON numbers come as float64).
func intArg(args map[string]any, key string) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	return 0
}

// stringOr returns the string arg value or a default.
func stringOr(args map[string]any, key, def string) string {
	if v, ok := args[key].(string); ok && v != "" {
		return v
	}
	return def
}

// setOptionalString sets a key in body if the arg is present and non-empty.
func setOptionalString(body map[string]any, args map[string]any, key string) {
	if v, ok := args[key].(string); ok && v != "" {
		body[key] = v
	}
}

// resolveIssueID takes a value that is either a UUID or an issue number string,
// and returns the UUID. If it's a number, it fetches the issue to get its ID.
func resolveIssueID(client *Client, slug, value string) (string, error) {
	// If it looks like a number, resolve it via the API
	num, err := strconv.Atoi(value)
	if err != nil {
		// Not a number — assume it's already a UUID
		return value, nil
	}

	data, err := client.get(fmt.Sprintf("/api/v1/projects/%s/issues/%d", slug, num), nil)
	if err != nil {
		return "", fmt.Errorf("fetching issue #%d: %w", num, err)
	}

	var issue struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &issue); err != nil {
		return "", fmt.Errorf("parsing issue response: %w", err)
	}
	if issue.ID == "" {
		return "", fmt.Errorf("issue #%d has no ID", num)
	}
	return issue.ID, nil
}
