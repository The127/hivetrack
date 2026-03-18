package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerIssueTools(s *server.MCPServer, client *Client) {
	s.AddTool(mcp.NewTool("list_issues",
		mcp.WithDescription("List issues in a project with optional filters"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithString("status", mcp.Description("Filter by status (e.g. todo, in_progress, in_review, done, cancelled)")),
		mcp.WithString("priority", mcp.Description("Filter by priority (none, low, medium, high, critical)")),
		mcp.WithString("type", mcp.Description("Filter by issue type (epic, task)")),
		mcp.WithString("text", mcp.Description("Full-text search in title/description")),
		mcp.WithString("triaged", mcp.Description("Filter by triaged status (true/false)")),
		mcp.WithString("backlog", mcp.Description("Filter backlog issues with no sprint (true/false)")),
		mcp.WithString("sprint_id", mcp.Description("Filter by sprint ID (UUID)")),
	), makeListIssues(client))

	s.AddTool(mcp.NewTool("get_issue",
		mcp.WithDescription("Get full details of a specific issue by project slug and issue number"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Issue number within the project")),
	), makeGetIssue(client))

	s.AddTool(mcp.NewTool("get_my_issues",
		mcp.WithDescription("Get all issues assigned to the current user across all projects"),
	), makeGetMyIssues(client))

	s.AddTool(mcp.NewTool("create_issue",
		mcp.WithDescription("Create a new issue in a project. Only title is required; everything else is optional."),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Issue title")),
		mcp.WithString("type", mcp.Description("Issue type: epic or task (default: task)")),
		mcp.WithString("priority", mcp.Description("Priority: none, low, medium, high, critical")),
		mcp.WithString("description", mcp.Description("Issue description (markdown)")),
		mcp.WithString("status", mcp.Description("Initial status (defaults to first status for archetype)")),
		mcp.WithString("estimate", mcp.Description("T-shirt size estimate: xs, s, m, l, xl")),
		mcp.WithString("sprint_id", mcp.Description("Sprint ID to assign to (UUID)")),
		mcp.WithString("milestone_id", mcp.Description("Milestone ID (UUID)")),
		mcp.WithString("parent_id", mcp.Description("Parent epic ID (UUID) — only for tasks")),
	), makeCreateIssue(client))

	s.AddTool(mcp.NewTool("update_issue",
		mcp.WithDescription("Update an existing issue. Only provide fields you want to change."),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Issue number")),
		mcp.WithString("title", mcp.Description("New title")),
		mcp.WithString("description", mcp.Description("New description (markdown)")),
		mcp.WithString("status", mcp.Description("New status")),
		mcp.WithString("priority", mcp.Description("New priority")),
		mcp.WithString("estimate", mcp.Description("New estimate (xs, s, m, l, xl)")),
		mcp.WithString("sprint_id", mcp.Description("Sprint ID (UUID), or 'null' to move to backlog")),
		mcp.WithString("milestone_id", mcp.Description("Milestone ID (UUID)")),
		mcp.WithString("parent_id", mcp.Description("Parent epic ID (UUID), or 'null' to remove parent")),
	), makeUpdateIssue(client))

	s.AddTool(mcp.NewTool("triage_issue",
		mcp.WithDescription("Triage an untriaged issue — set its initial status and optionally assign to sprint/milestone"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Issue number")),
		mcp.WithString("status", mcp.Required(), mcp.Description("Status to set")),
		mcp.WithString("sprint_id", mcp.Description("Sprint ID (UUID)")),
		mcp.WithString("milestone_id", mcp.Description("Milestone ID (UUID)")),
	), makeTriageIssue(client))
}

func makeListIssues(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		if slug == "" {
			return errResult(errMissing("slug")), nil
		}

		q := url.Values{}
		for _, key := range []string{"status", "priority", "type", "text", "triaged", "backlog", "sprint_id"} {
			if v, ok := args[key].(string); ok && v != "" {
				q.Set(key, v)
			}
		}

		data, err := client.get("/api/v1/projects/"+slug+"/issues", q)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatListIssues(data)), nil
	}
}

func makeGetIssue(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		number := intArg(args, "number")
		if slug == "" || number == 0 {
			return errResult(errMissing("slug, number")), nil
		}

		data, err := client.get(fmt.Sprintf("/api/v1/projects/%s/issues/%d", slug, number), nil)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatGetIssue(data)), nil
	}
}

func makeGetMyIssues(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := client.get("/api/v1/me/issues", nil)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatListIssues(data)), nil
	}
}

func makeCreateIssue(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		title, _ := args["title"].(string)
		if slug == "" || title == "" {
			return errResult(errMissing("slug, title")), nil
		}

		body := map[string]any{
			"title": title,
			"type":  stringOr(args, "type", "task"),
		}
		setOptionalString(body, args, "priority")
		setOptionalString(body, args, "description")
		setOptionalString(body, args, "status")
		setOptionalString(body, args, "estimate")
		setOptionalString(body, args, "sprint_id")
		setOptionalString(body, args, "milestone_id")
		setOptionalString(body, args, "parent_id")

		data, err := client.post("/api/v1/projects/"+slug+"/issues", body)
		if err != nil {
			return errResult(err), nil
		}

		result := formatCreateIssue(data, args)

		// Auto-triage when sprint_id is provided — assigning to a sprint implies triage.
		if sprintID, ok := args["sprint_id"].(string); ok && sprintID != "" {
			var created struct {
				Number int `json:"Number"`
			}
			if err := json.Unmarshal(data, &created); err == nil && created.Number > 0 {
				triageBody := map[string]any{"status": stringOr(args, "status", "todo")}
				triageBody["sprint_id"] = sprintID
				if milestoneID, ok := args["milestone_id"].(string); ok && milestoneID != "" {
					triageBody["milestone_id"] = milestoneID
				}
				_, triageErr := client.post(
					fmt.Sprintf("/api/v1/projects/%s/issues/%d/triage", slug, created.Number),
					triageBody,
				)
				if triageErr != nil {
					result += fmt.Sprintf("\n⚠ Auto-triage failed: %v", triageErr)
				} else {
					result += " ✓ triaged"
				}
			}
		}

		return textResult(result), nil
	}
}

func makeUpdateIssue(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		number := intArg(args, "number")
		if slug == "" || number == 0 {
			return errResult(errMissing("slug, number")), nil
		}

		body := map[string]any{}
		setOptionalString(body, args, "title")
		setOptionalString(body, args, "description")
		setOptionalString(body, args, "status")
		setOptionalString(body, args, "priority")
		setOptionalString(body, args, "estimate")

		// Handle nullable UUID fields — "null" string clears the field
		for _, key := range []string{"sprint_id", "milestone_id"} {
			if v, ok := args[key].(string); ok {
				if v == "null" {
					body[key] = nil
				} else if v != "" {
					body[key] = v
				}
			}
		}

		// parent_id: accept issue number (resolve to UUID) or UUID directly
		if v, ok := args["parent_id"].(string); ok {
			if v == "null" {
				body["parent_id"] = nil
			} else if v != "" {
				resolved, err := resolveIssueID(client, slug, v)
				if err != nil {
					return errResult(fmt.Errorf("resolving parent_id: %w", err)), nil
				}
				body["parent_id"] = resolved
			}
		}

		if len(body) == 0 {
			return errResult(fmt.Errorf("no fields to update")), nil
		}

		data, err := client.patch(fmt.Sprintf("/api/v1/projects/%s/issues/%d", slug, number), body)
		if err != nil {
			return errResult(err), nil
		}
		_ = data
		return textResult(formatUpdateIssue(number, args)), nil
	}
}

func makeTriageIssue(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		number := intArg(args, "number")
		status, _ := args["status"].(string)
		if slug == "" || number == 0 || status == "" {
			return errResult(errMissing("slug, number, status")), nil
		}

		body := map[string]any{
			"status": status,
		}
		setOptionalString(body, args, "sprint_id")
		setOptionalString(body, args, "milestone_id")

		_, err := client.post(fmt.Sprintf("/api/v1/projects/%s/issues/%d/triage", slug, number), body)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatTriageIssue(number, status, args)), nil
	}
}
