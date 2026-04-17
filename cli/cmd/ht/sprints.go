package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var sprintsCmd = &cli.Command{
	Name:  "sprints",
	Usage: "Manage sprints",
	Subcommands: []*cli.Command{
		sprintListCmd,
	},
}

var sprintListCmd = &cli.Command{
	Name:      "list",
	Usage:     "List sprints in a project",
	ArgsUsage: "<project-slug>",
	Action: func(c *cli.Context) error {
		slug := c.Args().First()
		if slug == "" {
			return cli.Exit("usage: ht sprints list <project-slug>", 1)
		}

		cl, err := mustClient(c)
		if err != nil {
			return err
		}

		sprints, err := cl.ListSprints(c.Context, slug)
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}

		if c.Bool("json") {
			return printJSON(c, sprints)
		}
		fmt.Fprintln(c.App.Writer, formatSprints(sprints))
		return nil
	},
}
