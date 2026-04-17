package main

import (
	"fmt"
	"strconv"

	htclient "github.com/the127/hivetrack/client"
	"github.com/urfave/cli/v2"
)

var issuesCmd = &cli.Command{
	Name:  "issues",
	Usage: "Manage issues",
	Subcommands: []*cli.Command{
		issueListCmd,
		issueShowCmd,
		issueCreateCmd,
		issueUpdateCmd,
		issueMeCmd,
	},
}

var issueListCmd = &cli.Command{
	Name:      "list",
	Usage:     "List issues in a project",
	ArgsUsage: "<project-slug>",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "status", Aliases: []string{"s"}, Usage: "Filter by status"},
		&cli.StringFlag{Name: "type", Aliases: []string{"t"}, Usage: "Filter by type (epic|task)"},
		&cli.StringFlag{Name: "priority", Aliases: []string{"p"}, Usage: "Filter by priority"},
		&cli.StringFlag{Name: "sprint", Usage: "Filter by sprint ID"},
		&cli.BoolFlag{Name: "backlog", Usage: "Show backlog issues (no sprint)"},
		&cli.BoolFlag{Name: "unrefined", Usage: "Show only unrefined issues"},
		&cli.IntFlag{Name: "limit", Aliases: []string{"l"}, Value: 50, Usage: "Max results"},
	},
	Action: func(c *cli.Context) error {
		slug := c.Args().First()
		if slug == "" {
			return cli.Exit("usage: ht issues list <project-slug>", 1)
		}

		cl, err := mustClient(c)
		if err != nil {
			return err
		}

		opts := htclient.ListIssuesOptions{
			Status:   c.String("status"),
			Type:     c.String("type"),
			Priority: c.String("priority"),
			SprintID: c.String("sprint"),
			Limit:    c.Int("limit"),
		}
		if c.Bool("backlog") {
			t := true
			opts.Backlog = &t
		}
		if c.Bool("unrefined") {
			f := false
			opts.Triaged = &f // untriaged is a proxy for unrefined in filter
		}

		items, total, err := cl.ListIssues(c.Context, slug, opts)
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}

		if c.Bool("json") {
			return printJSON(c, map[string]any{"items": items, "total": total})
		}
		fmt.Fprintln(c.App.Writer, formatIssueList(items, total))
		return nil
	},
}

var issueShowCmd = &cli.Command{
	Name:      "show",
	Usage:     "Show issue details",
	ArgsUsage: "<project-slug> <issue-number>",
	Action: func(c *cli.Context) error {
		slug := c.Args().Get(0)
		numStr := c.Args().Get(1)
		if slug == "" || numStr == "" {
			return cli.Exit("usage: ht issues show <project-slug> <issue-number>", 1)
		}
		number, err := strconv.Atoi(numStr)
		if err != nil {
			return cli.Exit("issue number must be an integer", 1)
		}

		cl, err := mustClient(c)
		if err != nil {
			return err
		}

		issue, err := cl.GetIssue(c.Context, slug, number)
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}

		if c.Bool("json") {
			return printJSON(c, issue)
		}
		fmt.Fprintln(c.App.Writer, formatIssue(issue))
		return nil
	},
}

var issueCreateCmd = &cli.Command{
	Name:      "create",
	Usage:     "Create a new issue",
	ArgsUsage: "<project-slug>",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "title", Aliases: []string{"T"}, Usage: "Issue title"},
		&cli.StringFlag{Name: "type", Aliases: []string{"t"}, Value: "task", Usage: "Issue type (epic|task)"},
		&cli.StringFlag{Name: "priority", Aliases: []string{"p"}, Usage: "Priority (none|low|medium|high|critical)"},
		&cli.StringFlag{Name: "estimate", Aliases: []string{"e"}, Usage: "Estimate (xs|s|m|l|xl)"},
		&cli.StringFlag{Name: "sprint", Usage: "Sprint ID"},
		&cli.StringFlag{Name: "milestone", Usage: "Milestone ID"},
		&cli.StringFlag{Name: "description", Aliases: []string{"d"}, Usage: "Issue description (markdown)"},
	},
	Action: func(c *cli.Context) error {
		slug := c.Args().First()
		if slug == "" {
			return cli.Exit("usage: ht issues create <project-slug> --title <title>", 1)
		}

		cl, err := mustClient(c)
		if err != nil {
			return err
		}

		if c.String("title") == "" {
			return cli.Exit("usage: ht issues create <project-slug> --title <title>", 1)
		}

		req := htclient.CreateIssueRequest{
			Title:       c.String("title"),
			Type:        c.String("type"),
			Priority:    c.String("priority"),
			Estimate:    c.String("estimate"),
			SprintID:    c.String("sprint"),
			MilestoneID: c.String("milestone"),
			Description: c.String("description"),
		}

		result, err := cl.CreateIssue(c.Context, slug, req)
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}

		if c.Bool("json") {
			return printJSON(c, result)
		}
		fmt.Fprintf(c.App.Writer, "Created #%d: %s\n", result.Number, req.Title)
		return nil
	},
}

var issueUpdateCmd = &cli.Command{
	Name:      "update",
	Usage:     "Update an issue",
	ArgsUsage: "<project-slug> <issue-number>",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "title", Aliases: []string{"T"}, Usage: "New title"},
		&cli.StringFlag{Name: "status", Aliases: []string{"s"}, Usage: "New status"},
		&cli.StringFlag{Name: "priority", Aliases: []string{"p"}, Usage: "New priority"},
		&cli.StringFlag{Name: "estimate", Aliases: []string{"e"}, Usage: "New estimate"},
		&cli.StringFlag{Name: "sprint", Usage: "Assign to sprint ID"},
		&cli.BoolFlag{Name: "clear-sprint", Usage: "Remove from sprint (move to backlog)"},
		&cli.StringFlag{Name: "milestone", Usage: "Assign to milestone ID"},
	},
	Action: func(c *cli.Context) error {
		slug := c.Args().Get(0)
		numStr := c.Args().Get(1)
		if slug == "" || numStr == "" {
			return cli.Exit("usage: ht issues update <project-slug> <issue-number> [flags]", 1)
		}
		number, err := strconv.Atoi(numStr)
		if err != nil {
			return cli.Exit("issue number must be an integer", 1)
		}

		cl, err := mustClient(c)
		if err != nil {
			return err
		}

		req := htclient.UpdateIssueRequest{}
		if c.IsSet("title") {
			req.Title = htclient.Set(c.String("title"))
		}
		if c.IsSet("status") {
			req.Status = htclient.Set(c.String("status"))
		}
		if c.IsSet("priority") {
			req.Priority = htclient.Set(c.String("priority"))
		}
		if c.IsSet("estimate") {
			req.Estimate = htclient.Set(c.String("estimate"))
		}
		if c.IsSet("milestone") {
			req.MilestoneID = htclient.Set(c.String("milestone"))
		}
		if c.Bool("clear-sprint") {
			req.SprintID = htclient.Null[string]()
		} else if c.IsSet("sprint") {
			req.SprintID = htclient.Set(c.String("sprint"))
		}

		if err := cl.UpdateIssue(c.Context, slug, number, req); err != nil {
			return cli.Exit(err.Error(), 1)
		}

		fmt.Fprintf(c.App.Writer, "Updated #%d.\n", number)
		return nil
	},
}

var issueMeCmd = &cli.Command{
	Name:  "me",
	Usage: "Issues assigned to or created by you",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "created", Aliases: []string{"c"}, Usage: "Show issues you created instead of issues assigned to you"},
	},
	Action: func(c *cli.Context) error {
		cl, err := mustClient(c)
		if err != nil {
			return err
		}

		var items []htclient.IssueSummary
		if c.Bool("created") {
			items, err = cl.GetMyCreatedIssues(c.Context)
		} else {
			items, err = cl.GetMyIssues(c.Context)
		}
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}

		if c.Bool("json") {
			return printJSON(c, items)
		}
		fmt.Fprintln(c.App.Writer, formatIssueList(items, len(items)))
		return nil
	},
}
