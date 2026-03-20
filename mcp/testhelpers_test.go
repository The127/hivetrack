package mcp

import (
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// extractText pulls the text content from a tool result's first content item.
func extractText(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	if tc, ok := result.Content[0].(mcp.TextContent); ok {
		return tc.Text
	}
	return ""
}

// contains reports whether substr is within s.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
