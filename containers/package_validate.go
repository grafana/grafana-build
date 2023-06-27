package containers

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

// NodeContainer returns a docker container with everything set up that is needed to build or run frontend tests.
func ValidatePackage(ctx context.Context, d *dagger.Client, file *dagger.File, src *dagger.Directory) (*dagger.Container, error) {
	// Get the node version from .nvmrc...
	nodeVersion, err := NodeVersion(d, src).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node version from source code: %w", err)
	}

	service := d.Container().From("alpine:latest").
		WithDirectory("/archive", ExtractedArchive(d, file)).
		WithWorkdir("/archive").
		WithExec([]string{"/archive/bin/grafana", "server"}).
		WithExposedPort(3000)

	return CypressContainer(d, CypressImage(nodeVersion)).
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithServiceBinding("grafana", service).
		WithEnvVariable("HOST", "grafana").
		WithEnvVariable("PORT", "3000").
		WithExec([]string{"yarn", "install", "--immutable"}).
		WithExec([]string{"/bin/sh", "-c", "/src/e2e/verify-release || true"}), nil
}
