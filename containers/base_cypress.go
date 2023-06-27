package containers

import (
	"dagger.io/dagger"
)

func CypressImage(version string) string {
	return "cypress/included:9.5.1"
}

// CypressContainer returns a docker container with everything set up that is needed to build or run e2e tests.
func CypressContainer(d *dagger.Client, base string) *dagger.Container {
	container := d.Container().From(base).WithEntrypoint([]string{})

	return container
}
