package containers

import (
	"fmt"
	"strings"

	"dagger.io/dagger"
)

func NodeImage(version string) string {
	return fmt.Sprintf("node:%s", strings.TrimPrefix(strings.TrimSpace(version), "v"))
}

// NodeContainer returns a docker container with everything set up that is needed to build or run frontend tests.
func NodeContainer(d *dagger.Client, base string) *dagger.Container {
	container := d.Container().From(base)

	// The Golang alpine containers don't come with make installed
	if strings.Contains(base, "alpine") {
		container = container.WithExec([]string{"apk", "update"})
		container = container.WithExec([]string{"apk", "add", "make"})
	}

	return container
}
