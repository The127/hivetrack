package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerCommentTools(s *server.MCPServer, client *Client) {
	s.AddTool(mcp.NewTool("list_comments",
		mcp.WithDescription("List comments on an issue"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Issue number")),
	), makeListComments(client))

	s.AddTool(mcp.NewTool("create_comment",
		mcp.WithDescription("Add a comment to an issue"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Issue number")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Comment body (markdown)")),
	), makeCreateComment(client))
}

func makeListComments(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		number := intArg(args, "number")
		if slug == "" || number == 0 {
			return errResult(errMissing("slug, number")), nil
		}

		data, err := client.get(fmt.Sprintf("/api/v1/projects/%s/issues/%d/comments", slug, number), nil)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatListComments(data)), nil
	}
}

func makeCreateComment(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		number := intArg(args, "number")
		body, _ := args["body"].(string)
		if slug == "" || number == 0 || body == "" {
			return errResult(errMissing("slug, number, body")), nil
		}

		_, err := client.post(fmt.Sprintf("/api/v1/projects/%s/issues/%d/comments", slug, number), map[string]any{
			"body": body,
		})
		if err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Added comment to #%d", number)), nil
	}
}
