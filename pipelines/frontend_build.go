package pipelines

import (
	"context"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

func GrafanaFrontendBuildDirectory(ctx context.Context, d *dagger.Client, src *dagger.Directory, nodeVersion string) (*dagger.Directory, error) {
	return containers.CompileFrontend(d, src, nodeVersion), nil
}
