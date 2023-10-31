package frontend

import "dagger.io/dagger"

func YarnInstall(c *dagger.Client, src *dagger.Directory, version string, cache *dagger.CacheVolume) *dagger.Container {
	return WithYarnCache(c.Container().From(NodeImage(version)), cache).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"yarn", "install", "--immutable"})
}