package containers

import (
	"dagger.io/dagger"
)

// Storybook returns a dagger.Directory which contains the built storybook server.
func Storybook(d *dagger.Client, src *dagger.Directory, opts *YarnCacheOpts, version, nodeVersion string) *dagger.Directory {
	c := NodeContainer(d, NodeImage(nodeVersion)).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src")

	c = WithYarnCache(c, opts)

	c = c.WithExec([]string{"yarn", "install", "--immutable"}).
		WithExec([]string{"yarn", "run", "storybook:build"})

	return c.Directory("./packages/grafana-ui/dist/storybook")
}
