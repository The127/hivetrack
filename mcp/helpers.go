package mcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	htclient "github.com/the127/hivetrack/client"
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

// parseUUIDList extracts a comma-separated list of UUIDs from an argument.
// Returns nil (not error) if the key is absent or empty.
func parseUUIDList(args map[string]any, key string) ([]string, error) {
	v, ok := args[key].(string)
	if !ok || v == "" {
		return nil, nil
	}
	parts := strings.Split(v, ",")
	ids := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Validate UUID format
		if _, err := uuid.Parse(p); err != nil {
			return nil, fmt.Errorf("invalid UUID %q: %w", p, err)
		}
		ids = append(ids, p)
	}
	return ids, nil
}

// resolveLabelNames fetches project labels and resolves comma-separated names to UUIDs.
// Returns nil if the key is absent or empty. Uses the typed client library.
func resolveLabelNames(client *Client, slug string, args map[string]any, key string) ([]string, error) {
	v, ok := args[key].(string)
	if !ok || v == "" {
		return nil, nil
	}
	return client.Typed().ResolveLabelNames(context.Background(), slug, v)
}

// fieldFromArgs extracts a string arg and returns it as a Field.
// If the value is "null", returns a Null field (explicit clear).
// If present and non-empty, returns a Set field.
// If absent or empty, returns an absent (zero) field.
func fieldFromArgs(args map[string]any, key string) htclient.Field[string] {
	if v, ok := args[key].(string); ok && v != "" {
		if v == "null" {
			return htclient.Null[string]()
		}
		return htclient.Set(v)
	}
	return htclient.Field[string]{}
}

// fieldFromArgsNoNull is like fieldFromArgs but does not treat "null" specially.
// Use for fields where clearing via "null" is not meaningful (e.g. sprint name, goal).
func fieldFromArgsNoNull(args map[string]any, key string) htclient.Field[string] {
	if v, ok := args[key].(string); ok && v != "" {
		return htclient.Set(v)
	}
	return htclient.Field[string]{}
}

// resolveIssueID takes a value that is either a UUID or an issue number string,
// and returns the UUID. If it's a number, it fetches the issue to get its ID.
func resolveIssueID(client *Client, slug, value string) (string, error) {
	num, err := strconv.Atoi(value)
	if err != nil {
		// Not a number — assume it's already a UUID
		return value, nil
	}

	issue, err := client.Typed().GetIssue(context.Background(), slug, num)
	if err != nil {
		return "", fmt.Errorf("fetching issue #%d: %w", num, err)
	}
	if issue.ID == "" {
		return "", fmt.Errorf("issue #%d has no ID", num)
	}
	return issue.ID, nil
}
