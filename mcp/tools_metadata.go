package mcp

import (
	"context"
	"fmt"
	"strings"

	htclient "github.com/the127/hivetrack/client"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerMetadataTools(s *server.MCPServer, client *Client) {
	s.AddTool(mcp.NewTool("list_labels",
		mcp.WithDescription("List all labels in a project"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
	), makeListLabels(client))

	s.AddTool(mcp.NewTool("create_label",
		mcp.WithDescription("Create a new label in a project"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Label name")),
		mcp.WithString("color", mcp.Required(), mcp.Description("Label color (hex, e.g. #ff0000)")),
	), makeCreateLabel(client))

	s.AddTool(mcp.NewTool("update_label",
		mcp.WithDescription("Update an existing label"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithString("label_id", mcp.Required(), mcp.Description("Label ID (UUID)")),
		mcp.WithString("name", mcp.Description("New label name")),
		mcp.WithString("color", mcp.Description("New label color (hex)")),
	), makeUpdateLabel(client))

	s.AddTool(mcp.NewTool("delete_label",
		mcp.WithDescription("Delete a label from a project"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithString("label_id", mcp.Required(), mcp.Description("Label ID (UUID)")),
	), makeDeleteLabel(client))

}

func makeListLabels(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug, _ := request.GetArguments()["slug"].(string)
		if slug == "" {
			return errResult(errMissing("slug")), nil
		}

		labels, err := client.Typed().ListLabels(ctx, slug)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatListLabels(labels)), nil
	}
}

func makeCreateLabel(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		name, _ := args["name"].(string)
		color, _ := args["color"].(string)
		if slug == "" || name == "" || color == "" {
			return errResult(errMissing("slug, name, color")), nil
		}

		id, err := client.Typed().CreateLabel(ctx, slug, name, color)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatCreateLabel(id, name, color)), nil
	}
}

func makeUpdateLabel(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		labelID, _ := args["label_id"].(string)
		if slug == "" || labelID == "" {
			return errResult(errMissing("slug, label_id")), nil
		}

		req := htclient.UpdateLabelRequest{}
		var changes []string
		if v, ok := args["name"].(string); ok && v != "" {
			req.Name = &v
			changes = append(changes, fmt.Sprintf("name → %s", v))
		}
		if v, ok := args["color"].(string); ok && v != "" {
			req.Color = &v
			changes = append(changes, fmt.Sprintf("color → %s", v))
		}
		if len(changes) == 0 {
			return errResult(fmt.Errorf("no fields to update")), nil
		}

		if err := client.Typed().UpdateLabel(ctx, slug, labelID, req); err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Updated label: %s", strings.Join(changes, ", "))), nil
	}
}

func makeDeleteLabel(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		labelID, _ := args["label_id"].(string)
		if slug == "" || labelID == "" {
			return errResult(errMissing("slug, label_id")), nil
		}

		if err := client.Typed().DeleteLabel(ctx, slug, labelID); err != nil {
			return errResult(err), nil
		}
		return textResult("Deleted label"), nil
	}
}

