package mcp

import (
	"context"
	"fmt"
	"strings"

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
		user, err := client.Typed().GetMe(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatCurrentUser(user)), nil
	}
}

func makeListUsers(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		users, err := client.Typed().ListUsers(ctx)
		if err != nil {
			return errResult(err), nil
		}
		if len(users) == 0 {
			return textResult("No users found."), nil
		}
		var sb strings.Builder
		for _, u := range users {
			admin := ""
			if u.IsAdmin {
				admin = " [admin]"
			}
			fmt.Fprintf(&sb, "• %s (%s)%s\n", u.DisplayName, u.Email, admin)
		}
		return textResult(sb.String()), nil
	}
}
