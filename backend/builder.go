package backend

import (
	"context"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/pipeline"
)

// Builder returns the container that is used to build the Grafana backend binaries.
// This container needs to have:
// * zig, for cross-compilation
// * golang
// * musl
func Builder(ctx context.Context, o *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	var (
		GoImage = "golang:1.21-alpine"
	)
	container := o.Client.Container(dagger.ContainerOpts{
		Platform: o.Platform,
	}).From(GoImage)

	return container, nil
}
