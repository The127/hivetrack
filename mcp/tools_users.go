package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerUserTools(s *server.MCPServer, client *Client) {
	s.AddTool(mcp.NewTool("list_users",
		mcp.WithDescription("List all users in the Hivetrack instance"),
	), makeListUsers(client))
}

func makeListUsers(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := client.get("/api/v1/users", nil)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(data), nil
	}
}
