package pipelines

import (
	"context"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
)

func GrafanaBackendBuildDirectory(ctx context.Context, d *dagger.Client, src *dagger.Directory, distro executil.Distribution, version string) (*dagger.Directory, error) {
	if distro == "" {
		return nil, fmt.Errorf("not a valid distribution")
	}

	buildinfo, err := containers.GetBuildInfo(ctx, d, src, version)
	if err != nil {
		return nil, err
	}

	return containers.CompileBackend(d, distro, src, buildinfo), nil
}

// GrafanaBackendBuild builds the Grafana backend.
func GrafanaBackendBuild(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	version := args.Context.String("version")
	distro := executil.Distribution(args.Context.String("distro"))
	output := filepath.Join("bin", string(distro))

	container, err := GrafanaBackendBuildDirectory(ctx, d, args.Grafana, distro, version)
	if err != nil {
		return err
	}

	if _, err := container.Export(ctx, output); err != nil {
		return err
	}

	return nil
}
