package containers

import (
	"fmt"
	"strings"

	"dagger.io/dagger"
)

// NPMPackages returns a dagger.Directory which contains the Grafana NPM packages from the grafana source code.
func NPMPackages(d *dagger.Client, platform dagger.Platform, src *dagger.Directory, opts *YarnCacheOpts, version, nodeVersion string) *dagger.Directory {
	c := NodeContainer(d, NodeImage(nodeVersion), platform).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src")

	c = WithYarnCache(c, opts)

	ersion := strings.TrimPrefix(version, "v")

	c = c.WithExec([]string{"mkdir", "npm-packages"}).
		WithExec([]string{"yarn", "install", "--immutable"}).
		WithExec([]string{"yarn", "run", "packages:build"}).
		// TODO: We should probably start reusing the yarn pnp map if we can figure that out instead of rerunning yarn install everywhere.
		WithExec([]string{"yarn", "run", "lerna", "version", ersion, "--exact", "--no-git-tag-version", "--no-push", "--force-publish", "-y"}).
		WithExec([]string{"yarn", "lerna", "exec", "--no-private", "--", "yarn", "pack", "--out", fmt.Sprintf("/src/npm-packages/%%s-%v.tgz", version)})

	return c.Directory("./npm-packages")
}
