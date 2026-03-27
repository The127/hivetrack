package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	htclient "github.com/the127/hivetrack/client"

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
		mcp.WithString("assignee_id", mcp.Description("Filter by assignee user ID (UUID)")),
		mcp.WithString("label_id", mcp.Description("Filter by label ID (UUID)")),
		mcp.WithString("label_name", mcp.Description("Filter by label name (resolved to ID)")),
		mcp.WithString("exclude_label_id", mcp.Description("Exclude issues with this label ID (UUID)")),
		mcp.WithString("exclude_label_name", mcp.Description("Exclude issues with this label name (resolved to ID)")),
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
		mcp.WithString("assignee_ids", mcp.Description("Comma-separated user IDs (UUIDs) to assign")),
		mcp.WithString("label_ids", mcp.Description("Comma-separated label IDs (UUIDs) to attach")),
		mcp.WithString("label_names", mcp.Description("Comma-separated label names (resolved to IDs). Alternative to label_ids.")),
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
		mcp.WithString("assignee_ids", mcp.Description("Comma-separated user IDs (UUIDs) to assign, or 'null' to clear all assignees")),
		mcp.WithString("label_ids", mcp.Description("Comma-separated label IDs (UUIDs), or 'null' to clear all labels")),
		mcp.WithString("label_names", mcp.Description("Comma-separated label names (resolved to IDs server-side). Alternative to label_ids.")),
		mcp.WithString("owner_id", mcp.Description("User ID (UUID) to set as owner, or 'null' to clear")),
	), makeUpdateIssue(client))

	s.AddTool(mcp.NewTool("add_checklist_item",
		mcp.WithDescription("Add a checklist item to a task"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Issue number")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Checklist item text")),
	), makeAddChecklistItem(client))

	s.AddTool(mcp.NewTool("update_checklist_item",
		mcp.WithDescription("Update a checklist item (toggle done or edit text)"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Issue number")),
		mcp.WithString("item_id", mcp.Required(), mcp.Description("Checklist item ID (UUID)")),
		mcp.WithString("text", mcp.Description("New text for the item")),
		mcp.WithBoolean("done", mcp.Description("Set completion status")),
	), makeUpdateChecklistItem(client))

	s.AddTool(mcp.NewTool("remove_checklist_item",
		mcp.WithDescription("Remove a checklist item from a task"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Issue number")),
		mcp.WithString("item_id", mcp.Required(), mcp.Description("Checklist item ID (UUID)")),
	), makeRemoveChecklistItem(client))

	s.AddTool(mcp.NewTool("triage_issue",
		mcp.WithDescription("Triage an untriaged issue — set its initial status and optionally assign to sprint/milestone, priority, and estimate in a single call"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Issue number")),
		mcp.WithString("status", mcp.Required(), mcp.Description("Status to set")),
		mcp.WithString("sprint_id", mcp.Description("Sprint ID (UUID)")),
		mcp.WithString("milestone_id", mcp.Description("Milestone ID (UUID)")),
		mcp.WithString("priority", mcp.Description("Priority: none, low, medium, high, critical")),
		mcp.WithString("estimate", mcp.Description("T-shirt size estimate: xs, s, m, l, xl")),
	), makeTriageIssue(client))

	s.AddTool(mcp.NewTool("delete_issue",
		mcp.WithDescription("Delete an issue permanently"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Issue number")),
	), makeDeleteIssue(client))

	s.AddTool(mcp.NewTool("refine_issue",
		mcp.WithDescription("Mark an issue as refined (ready for development)"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Issue number")),
	), makeRefineIssue(client))

	s.AddTool(mcp.NewTool("split_issue",
		mcp.WithDescription("Split an issue into multiple smaller issues"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Issue number")),
		mcp.WithString("titles", mcp.Required(), mcp.Description("Comma-separated titles for the new issues")),
	), makeSplitIssue(client))

	s.AddTool(mcp.NewTool("batch_update_issues",
		mcp.WithDescription("Apply the same field changes to multiple issues in one call"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithString("numbers", mcp.Required(), mcp.Description("Comma-separated issue numbers")),
		mcp.WithString("status", mcp.Description("New status")),
		mcp.WithString("priority", mcp.Description("New priority")),
		mcp.WithString("estimate", mcp.Description("New estimate (xs, s, m, l, xl)")),
		mcp.WithString("sprint_id", mcp.Description("Sprint ID (UUID), or 'null' to move to backlog")),
		mcp.WithString("milestone_id", mcp.Description("Milestone ID (UUID)")),
		mcp.WithString("assignee_ids", mcp.Description("Comma-separated user IDs (UUIDs), or 'null' to clear")),
		mcp.WithString("label_ids", mcp.Description("Comma-separated label IDs (UUIDs), or 'null' to clear")),
		mcp.WithString("label_names", mcp.Description("Comma-separated label names (resolved to IDs). Alternative to label_ids.")),
	), makeBatchUpdateIssues(client))

	s.AddTool(mcp.NewTool("add_issue_link",
		mcp.WithDescription("Add a link between two issues"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Project slug")),
		mcp.WithNumber("number", mcp.Required(), mcp.Description("Source issue number")),
		mcp.WithString("link_type", mcp.Required(), mcp.Description("Link type: blocks, is_blocked_by, duplicates, relates_to")),
		mcp.WithNumber("target_number", mcp.Required(), mcp.Description("Target issue number")),
	), makeAddIssueLink(client))
}

func makeListIssues(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		if slug == "" {
			return errResult(errMissing("slug")), nil
		}

		q := url.Values{}
		for _, key := range []string{"status", "priority", "type", "text", "triaged", "backlog", "sprint_id", "assignee_id", "label_id", "exclude_label_id"} {
			if v, ok := args[key].(string); ok && v != "" {
				q.Set(key, v)
			}
		}

		// Resolve label names to IDs if provided.
		if name, ok := args["label_name"].(string); ok && name != "" {
			if _, hasID := args["label_id"].(string); !hasID {
				ids, err := resolveLabelNames(client, slug, map[string]any{"n": name}, "n")
				if err != nil {
					return errResult(err), nil
				}
				if len(ids) > 0 {
					q.Set("label_id", ids[0])
				}
			}
		}
		if name, ok := args["exclude_label_name"].(string); ok && name != "" {
			if _, hasID := args["exclude_label_id"].(string); !hasID {
				ids, err := resolveLabelNames(client, slug, map[string]any{"n": name}, "n")
				if err != nil {
					return errResult(err), nil
				}
				if len(ids) > 0 {
					q.Set("exclude_label_id", ids[0])
				}
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
		if ids, err := parseUUIDList(args, "assignee_ids"); err != nil {
			return errResult(fmt.Errorf("invalid assignee_ids: %w", err)), nil
		} else if ids != nil {
			body["assignee_ids"] = ids
		}
		if ids, err := parseUUIDList(args, "label_ids"); err != nil {
			return errResult(fmt.Errorf("invalid label_ids: %w", err)), nil
		} else if ids != nil {
			body["label_ids"] = ids
		}
		// label_names: resolve names to UUIDs (only if label_ids not provided)
		if _, hasLabelIDs := body["label_ids"]; !hasLabelIDs {
			if ids, err := resolveLabelNames(client, slug, args, "label_names"); err != nil {
				return errResult(fmt.Errorf("resolving label names: %w", err)), nil
			} else if ids != nil {
				body["label_ids"] = ids
			}
		}

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

		// assignee_ids: comma-separated UUIDs, or "null" to clear
		if v, ok := args["assignee_ids"].(string); ok {
			if v == "null" {
				body["assignee_ids"] = []string{}
			} else if v != "" {
				ids, err := parseUUIDList(args, "assignee_ids")
				if err != nil {
					return errResult(fmt.Errorf("invalid assignee_ids: %w", err)), nil
				}
				if ids != nil {
					body["assignee_ids"] = ids
				}
			}
		}

		// label_ids: comma-separated UUIDs, or "null" to clear
		if v, ok := args["label_ids"].(string); ok {
			if v == "null" {
				body["label_ids"] = []string{}
			} else if v != "" {
				ids, err := parseUUIDList(args, "label_ids")
				if err != nil {
					return errResult(fmt.Errorf("invalid label_ids: %w", err)), nil
				}
				if ids != nil {
					body["label_ids"] = ids
				}
			}
		}

		// label_names: resolve names to UUIDs (only if label_ids not provided)
		if _, hasLabelIDs := args["label_ids"]; !hasLabelIDs {
			if ids, err := resolveLabelNames(client, slug, args, "label_names"); err != nil {
				return errResult(fmt.Errorf("resolving label names: %w", err)), nil
			} else if ids != nil {
				body["label_ids"] = ids
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

		// owner_id: UUID or "null" to clear
		if v, ok := args["owner_id"].(string); ok {
			if v == "null" {
				body["owner_id"] = nil
			} else if v != "" {
				body["owner_id"] = v
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

		req := htclient.TriageIssueRequest{
			Status: status,
		}
		if v, ok := args["sprint_id"].(string); ok && v != "" {
			req.SprintID = &v
		}
		if v, ok := args["milestone_id"].(string); ok && v != "" {
			req.MilestoneID = &v
		}
		if v, ok := args["priority"].(string); ok && v != "" {
			req.Priority = &v
		}
		if v, ok := args["estimate"].(string); ok && v != "" {
			req.Estimate = &v
		}

		if err := client.Typed().TriageIssue(ctx, slug, number, req); err != nil {
			return errResult(err), nil
		}
		return textResult(formatTriageIssue(number, status, args)), nil
	}
}

func makeAddChecklistItem(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		number := intArg(args, "number")
		text, _ := args["text"].(string)
		if slug == "" || number == 0 || text == "" {
			return errResult(errMissing("slug, number, text")), nil
		}

		data, err := client.post(fmt.Sprintf("/api/v1/projects/%s/issues/%d/checklist", slug, number), map[string]any{
			"text": text,
		})
		if err != nil {
			return errResult(err), nil
		}

		var resp struct {
			ID string `json:"ID"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			return textResult(fmt.Sprintf("Added checklist item to #%d: %q", number, text)), nil
		}
		return textResult(fmt.Sprintf("Added checklist item to #%d: %q (id: %s)", number, text, resp.ID)), nil
	}
}

func makeUpdateChecklistItem(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		number := intArg(args, "number")
		itemID, _ := args["item_id"].(string)
		if slug == "" || number == 0 || itemID == "" {
			return errResult(errMissing("slug, number, item_id")), nil
		}

		body := map[string]any{}
		if text, ok := args["text"].(string); ok && text != "" {
			body["text"] = text
		}
		if done, ok := args["done"].(bool); ok {
			body["done"] = done
		}
		if len(body) == 0 {
			return errResult(fmt.Errorf("provide text and/or done to update")), nil
		}

		_, err := client.patch(fmt.Sprintf("/api/v1/projects/%s/issues/%d/checklist/%s", slug, number, itemID), body)
		if err != nil {
			return errResult(err), nil
		}

		var changes []string
		if text, ok := body["text"].(string); ok {
			changes = append(changes, fmt.Sprintf("text → %q", text))
		}
		if done, ok := body["done"].(bool); ok {
			if done {
				changes = append(changes, "☑ done")
			} else {
				changes = append(changes, "☐ not done")
			}
		}
		return textResult(fmt.Sprintf("Updated checklist item on #%d: %s", number, strings.Join(changes, ", "))), nil
	}
}

func makeDeleteIssue(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		number := intArg(args, "number")
		if slug == "" || number == 0 {
			return errResult(errMissing("slug, number")), nil
		}

		if err := client.Typed().DeleteIssue(ctx, slug, number); err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Issue #%d deleted", number)), nil
	}
}

func makeRefineIssue(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		number := intArg(args, "number")
		if slug == "" || number == 0 {
			return errResult(errMissing("slug, number")), nil
		}

		if err := client.Typed().RefineIssue(ctx, slug, number); err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Issue #%d marked as refined", number)), nil
	}
}

func makeSplitIssue(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		number := intArg(args, "number")
		titlesStr, _ := args["titles"].(string)
		if slug == "" || number == 0 || titlesStr == "" {
			return errResult(errMissing("slug, number, titles")), nil
		}

		var titles []string
		for _, t := range strings.Split(titlesStr, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				titles = append(titles, t)
			}
		}

		data, err := client.post(fmt.Sprintf("/api/v1/projects/%s/issues/%d/split", slug, number), map[string]any{
			"titles": titles,
		})
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatSplitIssue(data)), nil
	}
}

func makeAddIssueLink(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		number := intArg(args, "number")
		linkType, _ := args["link_type"].(string)
		targetNumber := intArg(args, "target_number")
		if slug == "" || number == 0 || linkType == "" || targetNumber == 0 {
			return errResult(errMissing("slug, number, link_type, target_number")), nil
		}

		if err := client.Typed().AddIssueLink(ctx, slug, number, htclient.LinkType(linkType), targetNumber); err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Added link %s from #%d to #%d", linkType, number, targetNumber)), nil
	}
}

func makeBatchUpdateIssues(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		numbersStr, _ := args["numbers"].(string)
		if slug == "" || numbersStr == "" {
			return errResult(errMissing("slug, numbers")), nil
		}

		// Parse issue numbers.
		var numbers []int
		for _, s := range strings.Split(numbersStr, ",") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			n := 0
			for _, c := range s {
				if c < '0' || c > '9' {
					return errResult(fmt.Errorf("invalid issue number: %q", s)), nil
				}
				n = n*10 + int(c-'0')
			}
			numbers = append(numbers, n)
		}
		if len(numbers) == 0 {
			return errResult(fmt.Errorf("no valid issue numbers provided")), nil
		}

		// Build the batch update body for the backend endpoint.
		body := map[string]any{
			"numbers": numbers,
		}
		setOptionalString(body, args, "status")
		setOptionalString(body, args, "priority")
		setOptionalString(body, args, "estimate")

		// Handle sprint_id: "null" means clear.
		if v, ok := args["sprint_id"].(string); ok {
			if v == "null" {
				body["clear_sprint_id"] = true
			} else if v != "" {
				body["sprint_id"] = v
			}
		}
		setOptionalString(body, args, "milestone_id")

		if v, ok := args["assignee_ids"].(string); ok {
			if v == "null" {
				body["assignee_ids"] = []string{}
			} else if v != "" {
				ids, err := parseUUIDList(args, "assignee_ids")
				if err != nil {
					return errResult(fmt.Errorf("invalid assignee_ids: %w", err)), nil
				}
				if ids != nil {
					body["assignee_ids"] = ids
				}
			}
		}
		if v, ok := args["label_ids"].(string); ok {
			if v == "null" {
				body["label_ids"] = []string{}
			} else if v != "" {
				ids, err := parseUUIDList(args, "label_ids")
				if err != nil {
					return errResult(fmt.Errorf("invalid label_ids: %w", err)), nil
				}
				if ids != nil {
					body["label_ids"] = ids
				}
			}
		}
		if _, hasLabelIDs := body["label_ids"]; !hasLabelIDs {
			if ids, err := resolveLabelNames(client, slug, args, "label_names"); err != nil {
				return errResult(fmt.Errorf("resolving label names: %w", err)), nil
			} else if ids != nil {
				body["label_ids"] = ids
			}
		}

		data, err := client.post("/api/v1/projects/"+slug+"/issues/batch-update", body)
		if err != nil {
			return errResult(err), nil
		}
		var resp struct {
			Updated int `json:"Updated"`
		}
		if jsonErr := json.Unmarshal(data, &resp); jsonErr == nil {
			return textResult(fmt.Sprintf("Updated %d issue(s)", resp.Updated)), nil
		}
		return textResult(string(data)), nil
	}
}

func makeRemoveChecklistItem(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		number := intArg(args, "number")
		itemID, _ := args["item_id"].(string)
		if slug == "" || number == 0 || itemID == "" {
			return errResult(errMissing("slug, number, item_id")), nil
		}

		if err := client.Typed().RemoveChecklistItem(ctx, slug, number, itemID); err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Removed checklist item from #%d", number)), nil
	}
}
