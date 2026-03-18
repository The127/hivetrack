package mcp

import (
	"context"

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
}

func makeListProjects(client *Client) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := client.get("/api/v1/projects", nil)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(data), nil
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
		return jsonResult(data), nil
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
