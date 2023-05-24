package containers

import "github.com/grafana/grafana-build/cliutil"

type DockerOpts struct {
	// Registry is the docker Registry for the image.
	// If using '--save', then this will have no effect.
	// Uses docker hub by default.
	// Example: us.gcr.io/12345
	Registry string
	// Push will push the image to the container registry if it is true
	// If both Push and Save are true, then both will happen.
	Push bool
	// Save will save the image to the local filesystem.
	// If both Push and Save are true, then both will happen.
	Save bool

	// AlpineBase is supplied as a build-arg when building Grafana.
	// When building alpine versions of Grafana it uses this image as its base.
	AlpineBase string

	// UbuntuBase is supplied as a build-arg when building Grafana.
	// When building ubuntu versions of Grafana it uses this image as its base.
	UbuntuBase string
}

func DockerOptsFromFlags(c cliutil.CLIContext) *DockerOpts {
	return &DockerOpts{
		Registry:   c.String("registry"),
		Push:       c.Bool("push"),
		Save:       c.Bool("save"),
		AlpineBase: c.String("alpine-base"),
		UbuntuBase: c.String("ubuntu-base"),
	}
}
