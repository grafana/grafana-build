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

	// AlpineBaseARMv7 is supplied as a build-arg when building the Grafana docker image for the ARMv7 architecture.
	// When building alpine versions of Grafana it uses this image as its base.
	AlpineBaseARMv7 string

	// UbuntuBaseARMv7 is supplied as a build-arg when building the Grafana docker image for the ARMv7 architecture.
	// When building ubuntu versions of Grafana it uses this image as its base.
	UbuntuBaseARMv7 string

	// AlpineBaseARM64 is supplied as a build-arg when building the Grafana docker image for the ARM64 architecture.
	// When building alpine versions of Grafana it uses this image as its base.
	AlpineBaseARM64 string

	// UbuntuBaseARM64 is supplied as a build-arg when building the Grafana docker image for the ARM64 architecture.
	// When building ubuntu versions of Grafana it uses this image as its base.
	UbuntuBaseARM64 string
}

func DockerOptsFromFlags(c cliutil.CLIContext) *DockerOpts {
	return &DockerOpts{
		Registry:        c.String("registry"),
		AlpineBase:      c.String("alpine-base"),
		UbuntuBase:      c.String("ubuntu-base"),
		AlpineBaseARMv7: c.String("alpine-base-armv7"),
		UbuntuBaseARMv7: c.String("ubuntu-base-armv7"),
		AlpineBaseARM64: c.String("alpine-base-arm64"),
		UbuntuBaseARM64: c.String("ubuntu-base-arm64"),
	}
}
