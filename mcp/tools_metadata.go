package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerMetadataTools(s *server.MCPServer, client *Client) {
	s.AddTool(mcp.NewTool("list_labels",
		mcp.WithDescription("List all labels in a project"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID (UUID)")),
	), makeListLabels(client))

	s.AddTool(mcp.NewTool("list_milestones",
		mcp.WithDescription("List all milestones in a project"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID (UUID)")),
	), makeListMilestones(client))
}

func makeListLabels(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectID, _ := request.GetArguments()["project_id"].(string)
		if projectID == "" {
			return errResult(errMissing("project_id")), nil
		}

		data, err := client.get("/api/v1/projects/"+projectID+"/labels", nil)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(data), nil
	}
}

func makeListMilestones(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectID, _ := request.GetArguments()["project_id"].(string)
		if projectID == "" {
			return errResult(errMissing("project_id")), nil
		}

		data, err := client.get("/api/v1/projects/"+projectID+"/milestones", nil)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(data), nil
	}
}
