package containers

import (
	"log"

	"dagger.io/dagger"
)

// GolangContainer returns a dagger container with everything set up that is needed to build Grafana's Go backend or run the Golang tests.
func GolangContainer(d *dagger.Client, platform dagger.Platform, base string) *dagger.Container {
	log.Printf("Retrieving Go container based on `%s`", base)
	opts := dagger.ContainerOpts{
		Platform: platform,
	}

	container := d.Container(opts).From(base).
		WithExec([]string{"apk", "add", "zig", "--repository=https://dl-cdn.alpinelinux.org/alpine/edge/testing"}).
		WithExec([]string{"apk", "add", "--update", "build-base", "alpine-sdk", "musl", "musl-dev"})
	// Install the toolchain specifically for armv7 until we figure out why it's crashing w/ zig
	container = container.
		WithExec([]string{"mkdir", "/toolchain"}).
		WithExec([]string{"wget", "http://musl.cc/arm-linux-musleabihf-cross.tgz", "-P", "/toolchain"}).
		WithExec([]string{"tar", "-xvf", "/toolchain/arm-linux-musleabihf-cross.tgz", "-C", "/toolchain"})

	return container
}
