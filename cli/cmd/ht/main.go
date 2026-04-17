package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "ht",
		Usage:                "Hivetrack CLI",
		Version:              "0.1.0",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "Hivetrack server URL (overrides config)",
				EnvVars: []string{"HIVETRACK_SERVER"},
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output as JSON",
			},
		},
		Commands: []*cli.Command{
			loginCmd,
			logoutCmd,
			projectsCmd,
			issuesCmd,
			sprintsCmd,
			milestonesCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
