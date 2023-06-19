package containers

import "github.com/grafana/grafana-build/cliutil"

type DockerOpts struct {
	// Registry is the docker Registry for the image.
	// If using '--save', then this will have no effect.
	// Uses docker hub by default.
	// Example: us.gcr.io/12345
	Registry string

	// AlpineBase is supplied as a build-arg when building the Grafana docker image.
	// When building alpine versions of Grafana it uses this image as its base.
	AlpineBase string

	// UbuntuBase is supplied as a build-arg when building the Grafana docker image.
	// When building ubuntu versions of Grafana it uses this image as its base.
	UbuntuBase string
}

func DockerOptsFromFlags(c cliutil.CLIContext) *DockerOpts {
	return &DockerOpts{
		Registry:   c.String("registry"),
		AlpineBase: c.String("alpine-base"),
		UbuntuBase: c.String("ubuntu-base"),
	}
}
