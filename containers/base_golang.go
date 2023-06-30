package containers

import (
	"dagger.io/dagger"
)

// GolangContainer returns a dagger container with everything set up that is needed to build Grafana's Go backend or run the Golang tests.
func GolangContainer(d *dagger.Client, platform dagger.Platform, base string) *dagger.Container {
	opts := dagger.ContainerOpts{
		Platform: platform,
	}

	container := d.Container(opts).From(base).
		WithExec([]string{"apk", "add", "zig", "--repository=https://dl-cdn.alpinelinux.org/alpine/edge/testing"}).
		WithExec([]string{"apk", "add", "--update", "build-base", "alpine-sdk", "musl", "musl-dev"})

	return container
}
