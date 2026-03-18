package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewServer creates a configured MCP server with all Hivetrack tools registered.
func NewServer(client *Client) *server.MCPServer {
	s := server.NewMCPServer(
		"Hivetrack",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	registerProjectTools(s, client)
	registerIssueTools(s, client)
	registerSprintTools(s, client)
	registerMetadataTools(s, client)

	return s
}

// helper to build a text result from raw JSON
func jsonResult(data []byte) *mcp.CallToolResult {
	return mcp.NewToolResultText(string(data))
}

// helper to build an error result
func errResult(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}
