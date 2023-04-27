package containers

import (
	"strings"

	"dagger.io/dagger"
)

// GolangContainer returns a dagger container with everything set up that is needed to build Grafana's Go backend or run the Golang tests.
func GolangContainer(d *dagger.Client, base string) *dagger.Container {
	opts := dagger.ContainerOpts{}

	container := d.Container(opts).From(base)

	// The Golang alpine containers don't come with make or gcc installed
	if strings.Contains(base, "alpine") {
		container = container.WithExec([]string{"apk", "update"})
		container = container.WithExec([]string{"apk", "add", "make", "build-base"})
	}

	return container
}
