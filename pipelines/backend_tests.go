package pipelines

import (
	"context"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

// GrafanabackendTests runs the Grafana backend test containers for short (unit) and integration tests.
func GrafanaBackendTests(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	c := d.Pipeline("backend tests", dagger.PipelineOpts{
		Description: "Runs backend unit tests",
	})

	r := []*dagger.Container{
		containers.BackendTestShort(c, args.Path),
	}

	return containers.Run(ctx, r)
}

// GrafanabackendTestIntegration runs the Grafana backend test containers for short (unit) and integration tests.
func GrafanaBackendTestIntegration(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	c := d.Pipeline("integration tests", dagger.PipelineOpts{
		Description: "Runs backend integration tests",
	})

	r := []*dagger.Container{
		containers.BackendTestIntegration(c, args.Path),
	}

	return containers.Run(ctx, r)
}
