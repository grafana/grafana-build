package main

import (
	"context"
	"errors"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

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
		defer client.Close()

		args, err := pipelines.PipelineArgsFromContext(ctx, c)
		if err != nil {
			return err
		}

		if len(args.PackageInputOpts.Packages) == 0 {
			return errors.New("expected at least one package from a '--package' flag")
		}

		if err := pf(ctx, client, args); err != nil {
			return err
		}
		return nil
	}
}

func main() {
	ctx := context.Background()

	// TODO change the registerer if the user is running using a JSON file etc
	for k, v := range Artifacts {
		if err := globalCLI.Register(k, v); err != nil {
			panic(err)
		}
	}

	app := globalCLI.App()

	if err := app.RunContext(ctx, os.Args); err != nil {
		panic(err)
	}
}
