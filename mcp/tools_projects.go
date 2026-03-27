package mcp

import (
	"context"
	"fmt"
	"strings"

	htclient "github.com/the127/hivetrack/client"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerProjectTools(s *server.MCPServer, client *Client) {
	s.AddTool(mcp.NewTool("list_projects",
		mcp.WithDescription("List all projects the current user has access to"),
	), makeListProjects(client))

	s.AddTool(mcp.NewTool("get_project",
		mcp.WithDescription("Get details of a specific project by slug"),
		mcp.WithString("slug",
			mcp.Required(),
			mcp.Description("Project slug (URL-friendly identifier)"),
		),
	), makeGetProject(client))

	s.AddTool(mcp.NewTool("create_project",
		mcp.WithDescription("Create a new project. The creator is automatically added as project admin."),
		mcp.WithString("slug",
			mcp.Required(),
			mcp.Description("URL-friendly project identifier (lowercase, hyphens)"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Human-readable project name"),
		),
		mcp.WithString("archetype",
			mcp.Required(),
			mcp.Description("Project archetype: software or support"),
		),
		mcp.WithString("description",
			mcp.Description("Optional project description"),
		),
	), makeCreateProject(client))

	s.AddTool(mcp.NewTool("add_project_member",
		mcp.WithDescription("Add a user as a member of a project"),
		mcp.WithString("slug",
			mcp.Required(),
			mcp.Description("Project slug"),
		),
		mcp.WithString("user_id",
			mcp.Required(),
			mcp.Description("User ID (UUID) to add"),
		),
		mcp.WithString("role",
			mcp.Required(),
			mcp.Description("Role: project_admin, project_member, or viewer"),
		),
	), makeAddProjectMember(client))

	s.AddTool(mcp.NewTool("remove_project_member",
		mcp.WithDescription("Remove a user from a project"),
		mcp.WithString("slug",
			mcp.Required(),
			mcp.Description("Project slug"),
		),
		mcp.WithString("user_id",
			mcp.Required(),
			mcp.Description("User ID (UUID) to remove"),
		),
	), makeRemoveProjectMember(client))

	s.AddTool(mcp.NewTool("update_project",
		mcp.WithDescription("Update project settings. Only provide fields you want to change."),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID (UUID)")),
		mcp.WithString("name", mcp.Description("New project name")),
		mcp.WithString("description", mcp.Description("New project description")),
		mcp.WithBoolean("archived", mcp.Description("Archive or unarchive the project")),
		mcp.WithNumber("wip_limit_in_progress", mcp.Description("WIP limit for in_progress column (-1 to clear)")),
		mcp.WithNumber("wip_limit_in_review", mcp.Description("WIP limit for in_review column (-1 to clear)")),
	), makeUpdateProject(client))

	s.AddTool(mcp.NewTool("delete_project",
		mcp.WithDescription("Delete a project permanently"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID (UUID)")),
	), makeDeleteProject(client))
}

func makeListProjects(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projects, err := client.Typed().ListProjects(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatListProjects(projects)), nil
	}
}

func makeCreateProject(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		name, _ := args["name"].(string)
		archetype, _ := args["archetype"].(string)
		if slug == "" || name == "" || archetype == "" {
			return errResult(errMissing("slug, name, archetype")), nil
		}

		_, err := client.Typed().CreateProject(ctx, htclient.CreateProjectRequest{
			Slug:        slug,
			Name:        name,
			Archetype:   archetype,
			Description: stringOr(args, "description", ""),
		})
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatCreateProject(slug, name, archetype)), nil
	}
}

func makeAddProjectMember(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		userID, _ := args["user_id"].(string)
		role, _ := args["role"].(string)
		if slug == "" || userID == "" || role == "" {
			return errResult(errMissing("slug, user_id, role")), nil
		}

		htRole := htclient.ProjectRole(role)
		if err := client.Typed().AddProjectMember(ctx, slug, userID, htRole); err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Added %s to project %s as %s", userID, slug, role)), nil
	}
}

func makeRemoveProjectMember(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		slug, _ := args["slug"].(string)
		userID, _ := args["user_id"].(string)
		if slug == "" || userID == "" {
			return errResult(errMissing("slug, user_id")), nil
		}

		if err := client.Typed().RemoveProjectMember(ctx, slug, userID); err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Removed %s from project %s", userID, slug)), nil
	}
}

func makeUpdateProject(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		projectID, _ := args["project_id"].(string)
		if projectID == "" {
			return errResult(errMissing("project_id")), nil
		}

		req := htclient.UpdateProjectRequest{}
		hasChanges := false
		if v, ok := args["name"].(string); ok && v != "" {
			req.Name = &v
			hasChanges = true
		}
		if v, ok := args["description"].(string); ok && v != "" {
			req.Description = &v
			hasChanges = true
		}
		if v, ok := args["archived"].(bool); ok {
			req.Archived = &v
			hasChanges = true
		}
		if v, ok := args["wip_limit_in_progress"].(float64); ok {
			n := int(v)
			req.WipLimitInProgress = &n
			hasChanges = true
		}
		if v, ok := args["wip_limit_in_review"].(float64); ok {
			n := int(v)
			req.WipLimitInReview = &n
			hasChanges = true
		}

		if !hasChanges {
			return errResult(fmt.Errorf("no fields to update")), nil
		}

		if err := client.Typed().UpdateProject(ctx, projectID, req); err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Project %s updated", projectID)), nil
	}
}

func makeDeleteProject(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		projectID, _ := args["project_id"].(string)
		if projectID == "" {
			return errResult(errMissing("project_id")), nil
		}

		if err := client.Typed().DeleteProject(ctx, projectID); err != nil {
			return errResult(err), nil
		}
		return textResult(fmt.Sprintf("Project %s deleted", projectID)), nil
	}
}

func makeGetProject(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug, _ := request.GetArguments()["slug"].(string)
		if slug == "" {
			return errResult(errMissing("slug")), nil
		}

		project, err := client.Typed().GetProject(ctx, slug)
		if err != nil {
			return errResult(err), nil
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "%s (%s, %s)\n", project.Name, project.Slug, project.Archetype)
		fmt.Fprintf(&sb, "ID: %s\n", project.ID)
		if len(project.Members) > 0 {
			sb.WriteString("\nMembers:\n")
			for _, m := range project.Members {
				name := m.DisplayName
				if name == "" {
					name = m.Email
				}
				fmt.Fprintf(&sb, "  • %s (%s)\n", name, m.Role)
			}
		}
		return textResult(sb.String()), nil
	}
}
