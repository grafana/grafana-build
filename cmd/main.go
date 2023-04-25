package main

import (
	"log"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Commands: []*cli.Command{
		BackendCommands,
		PackageCommand,
		DebCommand,
		//RPMCommand,
		//MSICommand,
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

		v, err := args.DetectVersion(ctx, client, grafanaDir)
		if err != nil {
			return err
		}
		args.Version = v

		return pf(c.Context, client, grafanaDir, args)
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
