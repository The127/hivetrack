package mcp

import (
	"context"
	"strings"
	"time"

	htclient "github.com/the127/hivetrack/client"

	"github.com/mark3labs/mcp-go/mcp"
)

// testClient creates a Client pre-loaded with a non-expired test token.
func testClient(url string) *Client {
	provider := &htclient.StaticTokenProvider{
		Token: htclient.TokenCache{AccessToken: "tok", Expiry: time.Now().Add(time.Hour)},
	}
	return NewClient(url, provider)
}

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

// staticTokenProvider for raw client tests.
type staticTokenProvider struct{ tc htclient.TokenCache }

func (s staticTokenProvider) ProvideToken(_ context.Context) (htclient.TokenCache, error) {
	return s.tc, nil
}
