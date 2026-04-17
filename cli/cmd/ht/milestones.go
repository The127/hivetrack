package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var milestonesCmd = &cli.Command{
	Name:  "milestones",
	Usage: "Manage milestones",
	Subcommands: []*cli.Command{
		milestoneListCmd,
	},
}

var milestoneListCmd = &cli.Command{
	Name:      "list",
	Usage:     "List milestones in a project",
	ArgsUsage: "<project-slug>",
	Action: func(c *cli.Context) error {
		slug := c.Args().First()
		if slug == "" {
			return cli.Exit("usage: ht milestones list <project-slug>", 1)
		}

		cl, err := mustClient(c)
		if err != nil {
			return err
		}

		milestones, err := cl.ListMilestones(c.Context, slug)
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}

		if c.Bool("json") {
			return printJSON(c, milestones)
		}
		fmt.Fprintln(c.App.Writer, formatMilestones(milestones))
		return nil
	},
}
