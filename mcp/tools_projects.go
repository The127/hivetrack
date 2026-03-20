package mcp

import (
	"context"
	"fmt"

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
		data, err := client.get("/api/v1/projects", nil)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatListProjects(data)), nil
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

		body := map[string]any{
			"slug":      slug,
			"name":      name,
			"archetype": archetype,
		}
		setOptionalString(body, args, "description")

		data, err := client.post("/api/v1/projects", body)
		if err != nil {
			return errResult(err), nil
		}
		return textResult(formatCreateProject(data, slug, name, archetype)), nil
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

		body := map[string]any{
			"user_id": userID,
			"role":    role,
		}

		data, err := client.post("/api/v1/projects/"+slug+"/members", body)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(data), nil
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

		data, err := client.delete("/api/v1/projects/" + slug + "/members/" + userID)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(data), nil
	}
}

func makeUpdateProject(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		projectID, _ := args["project_id"].(string)
		if projectID == "" {
			return errResult(errMissing("project_id")), nil
		}

		body := map[string]any{}
		setOptionalString(body, args, "name")
		setOptionalString(body, args, "description")
		if v, ok := args["archived"].(bool); ok {
			body["archived"] = v
		}
		for _, key := range []string{"wip_limit_in_progress", "wip_limit_in_review"} {
			if v, ok := args[key].(float64); ok {
				if v == -1 {
					body[key] = nil
				} else {
					body[key] = int(v)
				}
			}
		}

		if len(body) == 0 {
			return errResult(fmt.Errorf("no fields to update")), nil
		}

		_, err := client.patch("/api/v1/projects/"+projectID, body)
		if err != nil {
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

		_, err := client.delete("/api/v1/projects/" + projectID)
		if err != nil {
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

		data, err := client.get("/api/v1/projects/"+slug, nil)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(data), nil
	}
}
