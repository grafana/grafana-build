package containers

import (
	"math/rand"
	"strconv"

	"dagger.io/dagger"
)

// NodeContainer returns a docker container with everything set up that is needed to build or run frontend tests.
func ValidatePackage(d *dagger.Client, service *dagger.Container, src *dagger.Directory, yarnCacheVolume *dagger.CacheVolume, nodeVersion string) *dagger.Directory {
	// The cypress container should never be cached
	r := rand.Int()
	c := CypressContainer(d, CypressImage(nodeVersion))

	c = WithYarnCache(c, &YarnCacheOpts{
		CacheVolume: yarnCacheVolume,
	})

	c = c.WithEnvVariable("CACHE", strconv.Itoa(r)).
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithServiceBinding("grafana", service).
		WithEnvVariable("HOST", "grafana").
		WithEnvVariable("PORT", "3000").
		WithExec([]string{"yarn", "install", "--immutable"}).
		WithExec([]string{"/bin/sh", "-c", "/src/e2e/verify-release"})

	return c.Directory("e2e/verify/specs")
}
