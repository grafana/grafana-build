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
}

func DockerOptsFromFlags(c cliutil.CLIContext) *DockerOpts {
	return &DockerOpts{
		Registry: c.String("registry"),
		Push:     c.Bool("push"),
		Save:     c.Bool("save"),
	}
}
