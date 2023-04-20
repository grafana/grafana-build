package main

import (
	"log"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:     "grafana",
			Usage:    "If set, initialize Grafana",
			Required: false,
			Value:    true,
		},
		&cli.StringFlag{
			Name:     "grafana-dir",
			Usage:    "Local Grafana dir to use, instead of git clone",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "grafana-ref",
			Usage:    "Grafana ref to clone, not valid if --grafana-dir is set",
			Required: false,
			Value:    "main",
		},
		&cli.BoolFlag{
			Name:  "enterprise",
			Usage: "If set, initialize Grafana Enterprise",
			Value: false,
		},
		&cli.StringFlag{
			Name:     "enterprise-dir",
			Usage:    "Local Grafana Enterprise dir to use, instead of git clone",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "enterprise-ref",
			Usage:    "Grafana Enterprise ref to clone, not valid if --enterprise-dir is set",
			Required: false,
			Value:    "main",
		},
		&cli.StringFlag{
			Name:     "build-id",
			Usage:    "Build ID to use, by default will be what is defined in package.json",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "github-token",
			Usage:    "Github token to use for git cloning, by default will be pulled from GitHub",
			Required: false,
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "Increase log verbosity",
			Value:   false,
		},
	},
	Commands: []*cli.Command{
		{
			Name:        "backend",
			Usage:       "Grafana Backend (Golang) operations",
			Subcommands: BackendCommands,
		},
		PackageCommand,
	},
}

func PipelineAction(pf pipelines.PipelineFunc) cli.ActionFunc {
	return func(c *cli.Context) error {
		var (
			ctx  = c.Context
			opts = []dagger.ClientOpt{}
		)
		if c.Bool("verbose") {
			opts = append(opts, dagger.WithLogOutput(os.Stderr))
		}
		client, err := dagger.Connect(ctx, opts...)
		if err != nil {
			return err
		}

		args, err := pipelines.PipelineArgsFromContext(c.Context, c)
		if err != nil {
			return err
		}

		grafanaDir, err := args.Grafana(ctx, client)
		if err != nil {
			return err
		}

		return pf(c.Context, client, grafanaDir, args)
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
