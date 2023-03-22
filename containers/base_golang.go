package containers

import (
	"dagger.io/dagger"
)

// GolangContainer returns a dagger container with everything set up that is needed to build Grafana's Go backend or run the Golang tests.
func GolangContainer(d *dagger.Client, platform dagger.Platform, base string) *dagger.Container {
	opts := dagger.ContainerOpts{}
	if platform != "" {
		opts.Platform = platform
	}

	container := d.Container(opts).From(base).
		WithExec([]string{
			"apt-get", "update", "-yq",
		}).
		WithExec([]string{
			"apt-get", "install", "musl",
		})

	return container
}
