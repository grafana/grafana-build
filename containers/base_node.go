package containers

import (
	"fmt"
	"strings"

	"dagger.io/dagger"
)

func NodeImage(version string) string {
	return fmt.Sprintf("node:%s-slim", strings.TrimPrefix(strings.TrimSpace(version), "v"))
}

// NodeContainer returns a docker container with everything set up that is needed to build or run frontend tests.
func NodeContainer(d *dagger.Client, base string) *dagger.Container {
	container := d.Container().From(base).
		WithExec([]string{"apt-get", "update", "-yq"}).
		WithExec([]string{"apt-get", "install", "-yq", "make"}).
		WithEnvVariable("NODE_OPTIONS", "--max_old_space_size=8000")

	return container
}
