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

	// Username is supplied to login to the docker registry when publishing images.
	Username string

	// Password is supplied to login to the docker registry when publishing images.
	Password string

	// Org overrides the organization when when publishing images.
	Org string

	// Latest is supplied to also tag as latest when publishing images.
	Latest bool
}

func DockerOptsFromFlags(c cliutil.CLIContext) *DockerOpts {
	return &DockerOpts{
		Registry:   c.String("registry"),
		AlpineBase: c.String("alpine-base"),
		UbuntuBase: c.String("ubuntu-base"),
		Username:   c.String("username"),
		Password:   c.String("password"),
		Org:        c.String("org"),
		Latest:     c.Bool("latest"),
	}
}
