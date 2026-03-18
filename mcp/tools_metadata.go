package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerMetadataTools(s *server.MCPServer, client *Client) {
	s.AddTool(mcp.NewTool("list_labels",
		mcp.WithDescription("List all labels in a project"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
	), makeListLabels(client))

	s.AddTool(mcp.NewTool("list_milestones",
		mcp.WithDescription("List all milestones in a project"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
	), makeListMilestones(client))
}

func makeListLabels(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug, _ := request.GetArguments()["slug"].(string)
		if slug == "" {
			return errResult(errMissing("slug")), nil
		}

		data, err := client.get("/api/v1/projects/"+slug+"/labels", nil)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(data), nil
	}
}

func makeListMilestones(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug, _ := request.GetArguments()["slug"].(string)
		if slug == "" {
			return errResult(errMissing("slug")), nil
		}

		data, err := client.get("/api/v1/projects/"+slug+"/milestones", nil)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(data), nil
	}
}
