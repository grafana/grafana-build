package backend

import (
	"dagger.io/dagger"
	"github.com/grafana/grafana-build/pipeline"
)

// Builder returns the container that is used to build the Grafana backend binaries.
// This container needs to have:
// * zig, for cross-compilation
// * golang
// * musl
func Builder(d *dagger.Client, platform dagger.Platform, args []pipeline.Argument) *dagger.Container {
	var (
		GoImage = "golang:1.21-alpine"
	)
	container := d.Container(dagger.ContainerOpts{
		Platform: platform,
	}).From(GoImage)

	return container
}
