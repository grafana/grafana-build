package frontend

import (
	"dagger.io/dagger"
)

func Build(builder *dagger.Container) *dagger.Directory {
	public := builder.
		WithExec([]string{"yarn", "run", "build"}).
		Directory("/src/public")

	return public
}

func BuildPlugins(builder *dagger.Container) *dagger.Directory {
	public := builder.
		WithExec([]string{"yarn", "install", "--immutable"}).
		WithExec([]string{"yarn", "run", "plugins:build-bundled"}).
		Directory("/src/plugins-bundled")

	return public
}

// WithYarnCache mounts the given YarnCacheDir in the provided container
func WithYarnCache(container *dagger.Container, vol *dagger.CacheVolume) *dagger.Container {
	yarnCacheDir := "/yarn/cache"
	c := container.WithEnvVariable("YARN_CACHE_FOLDER", yarnCacheDir)
	return c.WithMountedCache(yarnCacheDir, vol)
}
