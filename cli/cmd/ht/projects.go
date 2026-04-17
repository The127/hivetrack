package main

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
)

var projectsCmd = &cli.Command{
	Name:  "projects",
	Usage: "List projects",
	Action: func(c *cli.Context) error {
		cl, err := mustClient(c)
		if err != nil {
			return err
		}

		projects, err := cl.ListProjects(c.Context)
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}

		if c.Bool("json") {
			return printJSON(c, projects)
		}
		fmt.Fprintln(c.App.Writer, formatProjects(projects))
		return nil
	},
}

func printJSON(c *cli.Context, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return cli.Exit(fmt.Sprintf("json encoding failed: %v", err), 1)
	}
	fmt.Fprintln(c.App.Writer, string(data))
	return nil
}
