package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerMilestoneTools(s *server.MCPServer, client *Client) {
	s.AddTool(mcp.NewTool("list_milestones",
		mcp.WithDescription("List all milestones in a project with progress counts"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
	), makeListMilestones(client))

	s.AddTool(mcp.NewTool("create_milestone",
		mcp.WithDescription("Create a new milestone in a project"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Milestone title")),
		mcp.WithString("description", mcp.Description("Milestone description")),
		mcp.WithString("target_date", mcp.Description("Target date (RFC3339, e.g. 2026-06-30T00:00:00Z)")),
	), makeCreateMilestone(client))

	s.AddTool(mcp.NewTool("update_milestone",
		mcp.WithDescription("Update an existing milestone. Use close=true to close it, close=false to reopen."),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithString("milestone_id", mcp.Required(), mcp.Description("Milestone ID (UUID)")),
		mcp.WithString("title", mcp.Description("New title")),
		mcp.WithString("description", mcp.Description("New description")),
		mcp.WithString("target_date", mcp.Description("New target date (RFC3339)")),
		mcp.WithString("close", mcp.Description("Set to 'true' to close the milestone, 'false' to reopen")),
	), makeUpdateMilestone(client))

	s.AddTool(mcp.NewTool("delete_milestone",
		mcp.WithDescription("Delete a milestone from a project"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithString("milestone_id", mcp.Required(), mcp.Description("Milestone ID (UUID)")),
	), makeDeleteMilestone(client))
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
		return textResult(formatListMilestones(data)), nil
	}
}

func makeCreateMilestone(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		title, _ := args["title"].(string)
		if slug == "" || title == "" {
			return errResult(errMissing("slug, title")), nil
		}

		body := map[string]any{"title": title}
		setOptionalString(body, args, "description")
		setOptionalString(body, args, "target_date")

		data, err := client.post("/api/v1/projects/"+slug+"/milestones", body)
		if err != nil {
			return errResult(err), nil
		}

		var resp struct {
			ID string `json:"ID"`
		}
		if jsonErr := json.Unmarshal(data, &resp); jsonErr == nil && resp.ID != "" {
			return textResult(fmt.Sprintf("Created milestone %q (id: %s)", title, resp.ID)), nil
		}
		return textResult(fmt.Sprintf("Created milestone %q", title)), nil
	}
}

func makeUpdateMilestone(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		milestoneID, _ := args["milestone_id"].(string)
		if slug == "" || milestoneID == "" {
			return errResult(errMissing("slug, milestone_id")), nil
		}

		body := map[string]any{}
		setOptionalString(body, args, "title")
		setOptionalString(body, args, "description")
		setOptionalString(body, args, "target_date")
		if v, ok := args["close"].(string); ok && v != "" {
			body["close"] = v == "true"
		}
		if len(body) == 0 {
			return errResult(fmt.Errorf("no fields to update")), nil
		}

		_, err := client.patch(fmt.Sprintf("/api/v1/projects/%s/milestones/%s", slug, milestoneID), body)
		if err != nil {
			return errResult(err), nil
		}

		var changes []string
		for _, key := range []string{"title", "description", "target_date"} {
			if v, ok := body[key].(string); ok {
				changes = append(changes, fmt.Sprintf("%s → %s", key, v))
			}
		}
		if v, ok := body["close"].(bool); ok {
			if v {
				changes = append(changes, "closed")
			} else {
				changes = append(changes, "reopened")
			}
		}
		return textResult(fmt.Sprintf("Updated milestone: %s", strings.Join(changes, ", "))), nil
	}
}

func makeDeleteMilestone(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		milestoneID, _ := args["milestone_id"].(string)
		if slug == "" || milestoneID == "" {
			return errResult(errMissing("slug, milestone_id")), nil
		}

		_, err := client.delete(fmt.Sprintf("/api/v1/projects/%s/milestones/%s", slug, milestoneID))
		if err != nil {
			return errResult(err), nil
		}
		return textResult("Deleted milestone"), nil
	}
}
