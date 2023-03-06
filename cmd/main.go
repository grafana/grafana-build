package main

import (
	"log"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

func PipelineArgsFromContext(c *cli.Context) pipelines.PipelineArgs {
	args := pipelines.PipelineArgs{}
	args.Verbose = c.Bool("v")

	args.Path = c.Args().Get(0)
	if args.Path == "" {
		args.Path = "."
	}

	args.Context = c

	return args
}

func PipelineAction(pf pipelines.PipelineFunc) cli.ActionFunc {
	return func(c *cli.Context) error {
		var (
			args = PipelineArgsFromContext(c)
			ctx  = c.Context
			opts = []dagger.ClientOpt{}
		)

		if args.Verbose {
			opts = append(opts, dagger.WithLogOutput(os.Stderr))
		}

		client, err := dagger.Connect(ctx, opts...)
		if err != nil {
			return err
		}

		return pf(c.Context, client, args)
	}
}

var app = &cli.App{
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Value:   false,
		},
	},
	Commands: []*cli.Command{
		{
			Name:        "backend",
			Usage:       "Grafana Backend (Golang) operations",
			Subcommands: []*cli.Command{TestBackendUnit, TestBackendIntegration},
		},
	},
}

func main() {

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
