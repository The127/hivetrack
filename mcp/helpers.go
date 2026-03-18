package mcp

import "fmt"

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
