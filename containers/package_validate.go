package containers

import (
	"math/rand"
	"strconv"

	"dagger.io/dagger"
)

// NodeContainer returns a docker container with everything set up that is needed to build or run frontend tests.
func ValidatePackage(d *dagger.Client, file *dagger.File, src *dagger.Directory, nodeVersion string) (*dagger.Container, error) {
	service := d.Container().From("alpine:latest").
		WithDirectory("/archive", ExtractedArchive(d, file)).
		WithWorkdir("/archive").
		WithExec([]string{"/archive/bin/grafana", "server"}).
		WithExposedPort(3000)

	// The cypress container should never be cached
	r := rand.Int()

	return CypressContainer(d, CypressImage(nodeVersion)).
		WithEnvVariable("CACHE", strconv.Itoa(r)).
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithServiceBinding("grafana", service).
		WithEnvVariable("HOST", "grafana").
		WithEnvVariable("PORT", "3000").
		WithExec([]string{"yarn", "install", "--immutable"}).
		WithExec([]string{"/bin/sh", "-c", "/src/e2e/verify-release || true"}), nil
}
