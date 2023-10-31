package golang

import (
	"fmt"

	"dagger.io/dagger"
)

func DownloadURL(version, arch string) string {
	return fmt.Sprintf("https://go.dev/dl/go%s.linux-%s.tar.gz", version, arch)
}

func Container(d *dagger.Client, platform dagger.Platform, version string) *dagger.Container {
	opts := dagger.ContainerOpts{
		Platform: platform,
	}

	goImage := fmt.Sprintf("golang:%s-alpine", version)

	return d.Container(opts).From(goImage)
}

func WithCachedGoDependencies(container *dagger.Container, dir *dagger.Directory, cache *dagger.CacheVolume) *dagger.Container {
	return container.WithMountedCache("/go/pkg/mod", cache, dagger.ContainerWithMountedCacheOpts{
		Source: dir,
	})
}

func ModuleDir(d *dagger.Client, platform dagger.Platform, gomod, gosum *dagger.File, goVersion string) *dagger.Directory {
	container := Container(d, platform, goVersion).
		WithWorkdir("/src").
		WithMountedFile("/src/go.mod", gomod).
		WithMountedFile("/src/go.sum", gosum).
		WithExec([]string{"go", "mod", "download"})

	return container.Directory("/go/pkg/mod")
}
