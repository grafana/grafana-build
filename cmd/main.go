package main

import (
	"errors"
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
		RPMCommand,
		CDNCommand,
		WindowsInstallerCommand,
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

		grafanaDir, err := args.GrafanaOpts.Grafana(ctx, client)
		if err != nil {
			return err
		}

		v, err := args.GrafanaOpts.DetectVersion(ctx, client, grafanaDir)
		if err != nil {
			return err
		}
		args.GrafanaOpts.Version = v

		return pf(c.Context, client, grafanaDir, args)
	}
}

func PipelineActionWithPackageInput(pf pipelines.PipelineFuncWithPackageInput) cli.ActionFunc {
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

		if len(args.PackageInputOpts.Packages) == 0 {
			return errors.New("expected at least one package from a '--package' flag")
		}

		return pf(c.Context, client, args)
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
