package pipelines

import (
	"context"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
)

// GrafanaBackendBuild builds the Grafana backend.
func GrafanaBackendBuild(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	version := args.Context.String("version")
	distro := executil.Distribution(args.Context.String("distro"))
	if distro == "" {
		return fmt.Errorf("not a valid distribution")
	}

	buildinfo, err := containers.GetBuildInfo(ctx, d, args.Path, version)
	if err != nil {
		return err
	}

	output := filepath.Join("bin", string(distro))

	if _, err := containers.
		CompileBackend(d, distro, args.Path, buildinfo).
		Export(ctx, output); err != nil {
		panic(err)
	}

	return nil
}
