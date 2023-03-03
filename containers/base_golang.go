package containers

import (
	"strings"

	"dagger.io/dagger"
)

// GolangContainer returns a Golang container with everything set up that is needed to build or run tests.
func GolangContainer(d *dagger.Client, base string) *dagger.Container {
	container := d.Container().From(base)

	// The Golang alpine containers don't come with make installed
	if strings.Contains(base, "alpine") {
		container = container.WithExec([]string{"apk", "update"})
		container = container.WithExec([]string{"apk", "add", "make"})
	}

	return container
}
