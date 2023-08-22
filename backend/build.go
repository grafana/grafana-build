package backend

import (
	"context"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/pipeline"
)

// Build uses the given dagger.Container and runs the commands necessary to build the Grafana backend binaries.
func Build(ctx context.Context, d *dagger.Client, c *dagger.Container, opts *pipeline.BuildOpts) (*dagger.File, error) {
	return nil, nil
}
