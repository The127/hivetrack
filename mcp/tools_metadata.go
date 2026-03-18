package mcp

import (
	"context"
	"fmt"
	"strings"

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
		return textResult(formatListLabels(data)), nil
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

		data, err := client.post("/api/v1/projects/"+slug+"/labels", map[string]any{
			"name":  name,
			"color": color,
		})
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatCreateLabel(data, name, color)), nil
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

		body := map[string]any{}
		setOptionalString(body, args, "name")
		setOptionalString(body, args, "color")
		if len(body) == 0 {
			return errResult(fmt.Errorf("no fields to update")), nil
		}

		_, err := client.patch(fmt.Sprintf("/api/v1/projects/%s/labels/%s", slug, labelID), body)
		if err != nil {
			return errResult(err), nil
		}

		var changes []string
		if v, ok := body["name"].(string); ok {
			changes = append(changes, fmt.Sprintf("name → %s", v))
		}
		if v, ok := body["color"].(string); ok {
			changes = append(changes, fmt.Sprintf("color → %s", v))
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

		_, err := client.delete(fmt.Sprintf("/api/v1/projects/%s/labels/%s", slug, labelID))
		if err != nil {
			return errResult(err), nil
		}
		return textResult("Deleted label"), nil
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
