package mcp

import (
	"context"
	"fmt"
	"strings"

	htclient "github.com/the127/hivetrack/client"

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

		milestones, err := client.Typed().ListMilestones(ctx, slug)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatListMilestones(milestones)), nil
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

		id, err := client.Typed().CreateMilestone(ctx, slug, htclient.CreateMilestoneRequest{
			Title:       title,
			Description: stringOr(args, "description", ""),
			TargetDate:  stringOr(args, "target_date", ""),
		})
		if err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Created milestone %q (id: %s)", title, id)), nil
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

		optStr := func(key string) *string {
			if v, ok := args[key].(string); ok && v != "" {
				return &v
			}
			return nil
		}

		req := htclient.UpdateMilestoneRequest{
			Title:       optStr("title"),
			Description: optStr("description"),
			TargetDate:  optStr("target_date"),
		}
		if v, ok := args["close"].(string); ok && v != "" {
			b := v == "true"
			req.Close = &b
		}

		var changes []string
		if req.Title != nil {
			changes = append(changes, fmt.Sprintf("title → %s", *req.Title))
		}
		if req.Description != nil {
			changes = append(changes, fmt.Sprintf("description → %s", *req.Description))
		}
		if req.TargetDate != nil {
			changes = append(changes, fmt.Sprintf("target_date → %s", *req.TargetDate))
		}
		if req.Close != nil {
			if *req.Close {
				changes = append(changes, "closed")
			} else {
				changes = append(changes, "reopened")
			}
		}
		if len(changes) == 0 {
			return errResult(fmt.Errorf("no fields to update")), nil
		}

		if err := client.Typed().UpdateMilestone(ctx, slug, milestoneID, req); err != nil {
			return errResult(err), nil
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

		if err := client.Typed().DeleteMilestone(ctx, slug, milestoneID); err != nil {
			return errResult(err), nil
		}
		return textResult("Deleted milestone"), nil
	}
}
