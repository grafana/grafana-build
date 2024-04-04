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

	// TODO change the registerer if the user is running using a JSON file etc
	for k, v := range Artifacts {
		if err := globalCLI.Register(k, v); err != nil {
			panic(err)
		}
	}

	app := globalCLI.App()

	if err := app.RunContext(otel.FindParentTrace(ctx), os.Args); err != nil {
		if err := shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown tracer: %s", err.Error())
		}
		if err := shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown tracer: %s", err.Error())
		}
	}
}
