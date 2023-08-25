package main

import (
	"context"

	"dagger.io/dagger"
	"github.com/urfave/cli/v2"
)

func lintProject(ctx context.Context, dc *dagger.Client) error {
	workDir := dc.Host().Directory(".")
	container := dc.Container().
		From("golangci/golangci-lint:v1.52.2").
		WithWorkdir("/src").
		WithMountedDirectory("/src", workDir).
		WithExec([]string{"golangci-lint", "run"})

	if _, err := container.Sync(ctx); err != nil {
		return err
	}
	return nil
}

func lintAction(cliCtx *cli.Context) (rerr error) {
	dc, err := dagger.Connect(cliCtx.Context)
	if err != nil {
		return err
	}
	defer func() {
		if err := dc.Close(); rerr == nil && err != nil {
			rerr = err
		}
	}()
	return lintProject(cliCtx.Context, dc)
}
