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

	s.AddTool(mcp.NewTool("update_comment",
		mcp.WithDescription("Update the body of an existing comment"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Issue number")),
		mcp.WithString("comment_id", mcp.Required(), mcp.Description("Comment ID (UUID)")),
		mcp.WithString("body", mcp.Required(), mcp.Description("New comment body (markdown)")),
	), makeUpdateComment(client))

	s.AddTool(mcp.NewTool("delete_comment",
		mcp.WithDescription("Delete a comment from an issue"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Issue number")),
		mcp.WithString("comment_id", mcp.Required(), mcp.Description("Comment ID (UUID)")),
	), makeDeleteComment(client))
}

func makeListComments(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		number := intArg(args, "number")
		if slug == "" || number == 0 {
			return errResult(errMissing("slug, number")), nil
		}

		comments, total, err := client.Typed().ListComments(ctx, slug, number)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatListComments(comments, total)), nil
	}
}

func makeUpdateComment(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		number := intArg(args, "number")
		commentID, _ := args["comment_id"].(string)
		body, _ := args["body"].(string)
		if slug == "" || number == 0 || commentID == "" || body == "" {
			return errResult(errMissing("slug, number, comment_id, body")), nil
		}

		if err := client.Typed().UpdateComment(ctx, slug, number, commentID, body); err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Comment %s updated on #%d", commentID, number)), nil
	}
}

func makeDeleteComment(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		number := intArg(args, "number")
		commentID, _ := args["comment_id"].(string)
		if slug == "" || number == 0 || commentID == "" {
			return errResult(errMissing("slug, number, comment_id")), nil
		}

		if err := client.Typed().DeleteComment(ctx, slug, number, commentID); err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Comment %s deleted from #%d", commentID, number)), nil
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

		if err := client.Typed().CreateComment(ctx, slug, number, body); err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Added comment to #%d", number)), nil
	}
}
