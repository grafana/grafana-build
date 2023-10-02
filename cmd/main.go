package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/otel"
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:  "grafana-build",
	Usage: "A build tool for Grafana",
	Commands: []*cli.Command{
		BackendCommands,
		PackageCommand,
		DebCommand,
		RPMCommand,
		CDNCommand,
		DockerCommand,
		WindowsInstallerCommand,
		ZipCommand,
		ValidateCommand,
		ProImageCommand,
		StorybookCommand,
		NPMCommand,
		GCOMCommand,
	},
}

func PipelineAction(pf pipelines.PipelineFunc) cli.ActionFunc {
	return func(c *cli.Context) error {
		var (
			ctx  = c.Context
			opts = []dagger.ClientOpt{}
		)
		ctx, span := otel.Tracer("grafana-build").Start(ctx, fmt.Sprintf("pipeline-%s", c.Command.Name))
		defer span.End()

		if c.Bool("verbose") {
			opts = append(opts, dagger.WithLogOutput(os.Stderr))
		}
		client, err := dagger.Connect(ctx, opts...)
		if err != nil {
			otel.RecordFailed(span, err, "failed to connect to Dagger")
			return err
		}
		defer client.Close()

		args, err := pipelines.PipelineArgsFromContext(ctx, c)
		if err != nil {
			span.RecordError(err)
			otel.RecordFailed(span, err, "failed to load arguments")
			return err
		}

		grafanaDir, err := args.GrafanaOpts.Grafana(ctx, client)
		if err != nil {
			otel.RecordFailed(span, err, "failed to load grafana directory")
			return err
		}

		v, err := args.GrafanaOpts.DetectVersion(ctx, client, grafanaDir)
		if err != nil {
			otel.RecordFailed(span, err, "failed to detect version")
			return err
		}
		args.GrafanaOpts.Version = v
		pipelines.InjectPipelineArgsIntoSpan(span, args)

		if err := pf(ctx, client, grafanaDir, args); err != nil {
			otel.RecordFailed(span, err, "pipeline failed")
			return err
		}
		return nil
	}
}

func PipelineActionWithPackageInput(pf pipelines.PipelineFuncWithPackageInput) cli.ActionFunc {
	return func(c *cli.Context) error {
		var (
			ctx  = c.Context
			opts = []dagger.ClientOpt{}
		)
		ctx, span := otel.Tracer("grafana-build").Start(ctx, fmt.Sprintf("pipeline-%s", c.Command.Name))
		defer span.End()
		if c.Bool("verbose") {
			opts = append(opts, dagger.WithLogOutput(os.Stderr))
		}
		client, err := dagger.Connect(ctx, opts...)
		if err != nil {
			otel.RecordFailed(span, err, "failed to connect to Dagger")
			return err
		}
		defer client.Close()

		args, err := pipelines.PipelineArgsFromContext(ctx, c)
		if err != nil {
			otel.RecordFailed(span, err, "failed to load arguments")
			return err
		}

		pipelines.InjectPipelineArgsIntoSpan(span, args)

		if len(args.PackageInputOpts.Packages) == 0 {
			otel.RecordFailed(span, err, "no package provided")
			return errors.New("expected at least one package from a '--package' flag")
		}

		if err := pf(ctx, client, args); err != nil {
			otel.RecordFailed(span, err, "pipeline failed")
			return err
		}
		return nil
	}
}

func main() {
	ctx := context.Background()
	shutdown := otel.Setup(ctx)
	if err := app.RunContext(otel.FindParentTrace(ctx), os.Args); err != nil {
		if err := shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown tracer: %s", err.Error())
		}
		log.Fatal(err)
	}
	if err := shutdown(context.Background()); err != nil {
		log.Printf("Failed to shutdown tracer: %s", err.Error())
	}
}
