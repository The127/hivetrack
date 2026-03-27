package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerSprintTools(s *server.MCPServer, client *Client) {
	s.AddTool(mcp.NewTool("list_sprints",
		mcp.WithDescription("List all sprints in a project"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
	), makeListSprints(client))

	s.AddTool(mcp.NewTool("create_sprint",
		mcp.WithDescription("Create a new sprint in a project"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Sprint name")),
		mcp.WithString("goal", mcp.Description("Sprint goal")),
		mcp.WithString("start_date", mcp.Description("Start date (RFC3339, e.g. 2026-03-18T00:00:00Z)")),
		mcp.WithString("end_date", mcp.Description("End date (RFC3339, e.g. 2026-04-01T00:00:00Z)")),
	), makeCreateSprint(client))

	s.AddTool(mcp.NewTool("update_sprint",
		mcp.WithDescription("Update an existing sprint. When completing, warns about open issues unless force=true."),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithString("id", mcp.Required(), mcp.Description("Sprint ID (UUID)")),
		mcp.WithString("name", mcp.Description("New sprint name")),
		mcp.WithString("goal", mcp.Description("New sprint goal")),
		mcp.WithString("start_date", mcp.Description("New start date (RFC3339)")),
		mcp.WithString("end_date", mcp.Description("New end date (RFC3339)")),
		mcp.WithString("status", mcp.Description("New status: planning, active, completed")),
		mcp.WithBoolean("force", mcp.Description("Force sprint completion even if open issues remain (they will be moved to backlog)")),
		mcp.WithString("move_to_sprint_id", mcp.Description("When completing, move open issues to this sprint ID instead of backlog")),
	), makeUpdateSprint(client))

	s.AddTool(mcp.NewTool("delete_sprint",
		mcp.WithDescription("Delete a sprint permanently"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithString("sprint_id", mcp.Required(), mcp.Description("Sprint ID (UUID)")),
	), makeDeleteSprint(client))

	s.AddTool(mcp.NewTool("get_sprint_burndown",
		mcp.WithDescription("Get the burndown chart data for a sprint"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithString("sprint_id", mcp.Required(), mcp.Description("Sprint ID (UUID)")),
	), makeGetSprintBurndown(client))
}

func makeListSprints(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug, _ := request.GetArguments()["slug"].(string)
		if slug == "" {
			return errResult(errMissing("slug")), nil
		}

		data, err := client.get("/api/v1/projects/"+slug+"/sprints", nil)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatListSprints(data)), nil
	}
}

func makeCreateSprint(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		name, _ := args["name"].(string)
		if slug == "" || name == "" {
			return errResult(errMissing("slug, name")), nil
		}

		body := map[string]any{
			"name": name,
		}
		setOptionalString(body, args, "goal")
		setOptionalString(body, args, "start_date")
		setOptionalString(body, args, "end_date")

		data, err := client.post("/api/v1/projects/"+slug+"/sprints", body)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatCreateSprint(data, name)), nil
	}
}

func makeDeleteSprint(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		sprintID, _ := args["sprint_id"].(string)
		if slug == "" || sprintID == "" {
			return errResult(errMissing("slug, sprint_id")), nil
		}

		_, err := client.delete(fmt.Sprintf("/api/v1/projects/%s/sprints/%s", slug, sprintID))
		if err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Sprint %s deleted", sprintID)), nil
	}
}

func makeGetSprintBurndown(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		sprintID, _ := args["sprint_id"].(string)
		if slug == "" || sprintID == "" {
			return errResult(errMissing("slug, sprint_id")), nil
		}

		data, err := client.get(fmt.Sprintf("/api/v1/projects/%s/sprints/%s/burndown", slug, sprintID), nil)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatSprintBurndown(data)), nil
	}
}

func makeUpdateSprint(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		id, _ := args["id"].(string)
		if slug == "" || id == "" {
			return errResult(errMissing("slug, id")), nil
		}

		body := map[string]any{}
		setOptionalString(body, args, "name")
		setOptionalString(body, args, "goal")
		setOptionalString(body, args, "start_date")
		setOptionalString(body, args, "end_date")
		setOptionalString(body, args, "status")

		// Pass force flag to backend (for sprint completion with open issues).
		if force, ok := args["force"].(bool); ok && force {
			body["force"] = true
		}

		// Map move_to_sprint_id to the backend field name.
		if moveID, ok := args["move_to_sprint_id"].(string); ok && moveID != "" {
			body["move_open_issues_to_sprint_id"] = moveID
		}

		if len(body) == 0 {
			return errResult(fmt.Errorf("no fields to update")), nil
		}

		data, err := client.patch(fmt.Sprintf("/api/v1/projects/%s/sprints/%s", slug, id), body)
		if err != nil {
			return errResult(err), nil
		}
		_ = data
		return textResult(formatUpdateSprint(args)), nil
	}
}
