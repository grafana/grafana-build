package containers

import "dagger.io/dagger"

func WithCachedGoDependencies(container *dagger.Container, dir *dagger.Directory, cache *dagger.CacheVolume) *dagger.Container {
	return container.WithMountedCache("/go/pkg/mod", cache, dagger.ContainerWithMountedCacheOpts{
		Source: dir,
	})
}

func DownloadGolangDependencies(d *dagger.Client, platform dagger.Platform, gomod, gosum *dagger.File) *dagger.Directory {
	container := GolangContainer(d, platform, GoImage).
		WithWorkdir("/src").
		WithMountedFile("/src/go.mod", gomod).
		WithMountedFile("/src/go.sum", gosum).
		WithExec([]string{"go", "mod", "download"})

	return container.Directory("/go/pkg/mod")
}
