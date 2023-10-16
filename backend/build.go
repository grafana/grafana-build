package backend

import (
	"context"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/pipeline"
)

// Build uses the given dagger.Container and runs the commands necessary to build the Grafana backend binaries.
// It returns the `bin` directory with the compiled binaries in it.
func Build(ctx context.Context, o *pipeline.ArtifactBuildOpts) (*dagger.Directory, error) {
	builder := o.Builder

	return builder.WithExec([]string{"mkdir", "-p", "bin"}).Directory("./bin"), nil
}
