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

	s.AddTool(mcp.NewTool("get_current_user",
		mcp.WithDescription("Get the currently authenticated user's profile"),
	), makeGetCurrentUser(client))
}

func makeGetCurrentUser(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := client.get("/api/v1/users/me", nil)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatCurrentUser(data)), nil
	}
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
