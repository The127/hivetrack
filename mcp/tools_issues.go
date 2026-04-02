package mcp

import (
	"context"
	"fmt"
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
		mcp.WithString("on_hold", mcp.Description("Filter by on-hold status (true/false)")),
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
		mcp.WithString("on_hold", mcp.Description("Set on-hold status: 'true' or 'false'")),
		mcp.WithString("hold_reason", mcp.Description("Hold reason: waiting_on_customer, waiting_on_external, blocked_by_issue")),
		mcp.WithString("hold_note", mcp.Description("Optional note about why the issue is on hold")),
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
		mcp.WithString("on_hold", mcp.Description("Set on-hold: 'true' or 'false'")),
		mcp.WithString("hold_reason", mcp.Description("Hold reason: waiting_on_customer, waiting_on_external, blocked_by_issue")),
		mcp.WithString("hold_note", mcp.Description("Optional hold note")),
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

		opts := htclient.ListIssuesOptions{
			Status:         stringOr(args, "status", ""),
			Priority:       stringOr(args, "priority", ""),
			Type:           stringOr(args, "type", ""),
			Text:           stringOr(args, "text", ""),
			SprintID:       stringOr(args, "sprint_id", ""),
			AssigneeID:     stringOr(args, "assignee_id", ""),
			LabelID:        stringOr(args, "label_id", ""),
			ExcludeLabelID: stringOr(args, "exclude_label_id", ""),
		}
		if v, ok := args["triaged"].(string); ok && v != "" {
			b := v == "true"
			opts.Triaged = &b
		}
		if v, ok := args["backlog"].(string); ok && v != "" {
			b := v == "true"
			opts.Backlog = &b
		}
		if v, ok := args["on_hold"].(string); ok && v != "" {
			b := v == "true"
			opts.OnHold = &b
		}

		// Resolve label names to IDs if provided.
		if name, ok := args["label_name"].(string); ok && name != "" && opts.LabelID == "" {
			ids, err := client.Typed().ResolveLabelNames(ctx, slug, name)
			if err != nil {
				return errResult(err), nil
			}
			if len(ids) > 0 {
				opts.LabelID = ids[0]
			}
		}
		if name, ok := args["exclude_label_name"].(string); ok && name != "" && opts.ExcludeLabelID == "" {
			ids, err := client.Typed().ResolveLabelNames(ctx, slug, name)
			if err != nil {
				return errResult(err), nil
			}
			if len(ids) > 0 {
				opts.ExcludeLabelID = ids[0]
			}
		}

		items, total, err := client.Typed().ListIssues(ctx, slug, opts)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatListIssues(items, total)), nil
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

		issue, err := client.Typed().GetIssue(ctx, slug, number)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatGetIssue(issue)), nil
	}
}

func makeGetMyIssues(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		items, err := client.Typed().GetMyIssues(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatListIssues(items, len(items))), nil
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

		req := htclient.CreateIssueRequest{
			Title:       title,
			Type:        stringOr(args, "type", "task"),
			Priority:    stringOr(args, "priority", ""),
			Description: stringOr(args, "description", ""),
			Status:      stringOr(args, "status", ""),
			Estimate:    stringOr(args, "estimate", ""),
			SprintID:    stringOr(args, "sprint_id", ""),
			MilestoneID: stringOr(args, "milestone_id", ""),
			ParentID:    stringOr(args, "parent_id", ""),
		}
		if ids, err := parseUUIDList(args, "assignee_ids"); err != nil {
			return errResult(fmt.Errorf("invalid assignee_ids: %w", err)), nil
		} else if ids != nil {
			req.AssigneeIDs = ids
		}
		if ids, err := parseUUIDList(args, "label_ids"); err != nil {
			return errResult(fmt.Errorf("invalid label_ids: %w", err)), nil
		} else if ids != nil {
			req.LabelIDs = ids
		}
		if req.LabelIDs == nil {
			if ids, err := resolveLabelNames(client, slug, args, "label_names"); err != nil {
				return errResult(fmt.Errorf("resolving label names: %w", err)), nil
			} else if ids != nil {
				req.LabelIDs = ids
			}
		}

		created, err := client.Typed().CreateIssue(ctx, slug, req)
		if err != nil {
			return errResult(err), nil
		}

		result := formatCreateIssue(created, args)

		// Auto-triage when sprint_id is provided — assigning to a sprint implies triage.
		if sprintID, ok := args["sprint_id"].(string); ok && sprintID != "" {
			triageReq := htclient.TriageIssueRequest{Status: stringOr(args, "status", "todo")}
			triageReq.SprintID = &sprintID
			if milestoneID, ok := args["milestone_id"].(string); ok && milestoneID != "" {
				triageReq.MilestoneID = &milestoneID
			}
			if triageErr := client.Typed().TriageIssue(ctx, slug, created.Number, triageReq); triageErr != nil {
				result += fmt.Sprintf("\n⚠ Auto-triage failed: %v", triageErr)
			} else {
				result += " ✓ triaged"
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

		req := htclient.UpdateIssueRequest{}

		// Helper: parse string arg as Set or Null field.
		setOrNull := func(key string) htclient.Field[string] {
			if v, ok := args[key].(string); ok && v != "" {
				if v == "null" {
					return htclient.Null[string]()
				}
				return htclient.Set(v)
			}
			return htclient.Field[string]{}
		}

		req.Title = setOrNull("title")
		req.Description = setOrNull("description")
		req.Status = setOrNull("status")
		req.Priority = setOrNull("priority")
		req.Estimate = setOrNull("estimate")
		req.SprintID = setOrNull("sprint_id")
		req.MilestoneID = setOrNull("milestone_id")
		req.OwnerID = setOrNull("owner_id")

		if v, ok := args["parent_id"].(string); ok && v != "" {
			if v == "null" {
				req.ParentID = htclient.Null[string]()
			} else {
				resolved, err := resolveIssueID(client, slug, v)
				if err != nil {
					return errResult(fmt.Errorf("resolving parent_id: %w", err)), nil
				}
				req.ParentID = htclient.Set(resolved)
			}
		}

		if v, ok := args["assignee_ids"].(string); ok && v != "" {
			if v == "null" {
				req.AssigneeIDs = htclient.Null[[]string]()
			} else {
				ids, err := parseUUIDList(args, "assignee_ids")
				if err != nil {
					return errResult(fmt.Errorf("invalid assignee_ids: %w", err)), nil
				}
				if ids != nil {
					req.AssigneeIDs = htclient.Set(ids)
				}
			}
		}

		if v, ok := args["label_ids"].(string); ok && v != "" {
			if v == "null" {
				req.LabelIDs = htclient.Null[[]string]()
			} else {
				ids, err := parseUUIDList(args, "label_ids")
				if err != nil {
					return errResult(fmt.Errorf("invalid label_ids: %w", err)), nil
				}
				if ids != nil {
					req.LabelIDs = htclient.Set(ids)
				}
			}
		}
		if req.LabelIDs.IsAbsent() {
			if ids, err := resolveLabelNames(client, slug, args, "label_names"); err != nil {
				return errResult(fmt.Errorf("resolving label names: %w", err)), nil
			} else if ids != nil {
				req.LabelIDs = htclient.Set(ids)
			}
		}

		if v, ok := args["on_hold"].(string); ok && v != "" {
			req.OnHold = htclient.Set(v == "true")
		}
		req.HoldReason = setOrNull("hold_reason")
		req.HoldNote = setOrNull("hold_note")

		if err := client.Typed().UpdateIssue(ctx, slug, number, req); err != nil {
			return errResult(err), nil
		}
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

		id, err := client.Typed().AddChecklistItem(ctx, slug, number, text)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Added checklist item to #%d: %q (id: %s)", number, text, id)), nil
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

		req := htclient.UpdateChecklistItemRequest{}
		var changes []string
		if text, ok := args["text"].(string); ok && text != "" {
			req.Text = &text
			changes = append(changes, fmt.Sprintf("text → %q", text))
		}
		if done, ok := args["done"].(bool); ok {
			req.Done = &done
			if done {
				changes = append(changes, "☑ done")
			} else {
				changes = append(changes, "☐ not done")
			}
		}
		if len(changes) == 0 {
			return errResult(fmt.Errorf("provide text and/or done to update")), nil
		}

		if err := client.Typed().UpdateChecklistItem(ctx, slug, number, itemID, req); err != nil {
			return errResult(err), nil
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

		result, err := client.Typed().SplitIssue(ctx, slug, number, titles)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatSplitIssue(result)), nil
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

		batchReq := htclient.BatchUpdateIssuesRequest{Numbers: numbers}
		setFromMap := func(m map[string]any, key string) htclient.Field[string] {
			if v, ok := m[key].(string); ok && v != "" {
				return htclient.Set(v)
			}
			return htclient.Field[string]{}
		}
		batchReq.Status = setFromMap(body, "status")
		batchReq.Priority = setFromMap(body, "priority")
		batchReq.Estimate = setFromMap(body, "estimate")
		batchReq.SprintID = setFromMap(body, "sprint_id")
		batchReq.MilestoneID = setFromMap(body, "milestone_id")
		if body["clear_sprint_id"] == true {
			batchReq.ClearSprintID = htclient.Set(true)
		}
		if ids, ok := body["assignee_ids"].([]string); ok {
			batchReq.AssigneeIDs = htclient.Set(ids)
		}
		if ids, ok := body["label_ids"].([]string); ok {
			batchReq.LabelIDs = htclient.Set(ids)
		}
		if v, ok := args["on_hold"].(string); ok && v != "" {
			batchReq.OnHold = htclient.Set(v == "true")
		}
		batchReq.HoldReason = setFromMap(body, "hold_reason")
		batchReq.HoldNote = setFromMap(body, "hold_note")
		result, err := client.Typed().BatchUpdateIssues(ctx, slug, batchReq)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Updated %d issue(s)", result.Updated)), nil
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
