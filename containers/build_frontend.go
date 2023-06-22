package containers

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

// NodeVersionContainer returns a container whose `stdout` will return the node version from the '.nvmrc' file in the directory 'src'.
func NodeVersion(d *dagger.Client, src *dagger.Directory) *dagger.Container {
	return d.Container().From("alpine:3.17").
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"cat", ".nvmrc"})
}

func CompileFrontend(d *dagger.Client, src *dagger.Directory, opts *YarnCacheOpts, nodeVersion string) *dagger.Directory {
	c := NodeContainer(d, NodeImage(nodeVersion)).
		WithDirectory("/src", src).
		WithWorkdir("/src")

	c = WithYarnCache(c, opts)

	// Get the node version from the 'src' directories '.nvmrc' file.
	public := c.
		WithExec([]string{"yarn", "install", "--immutable"}).
		WithExec([]string{"yarn", "run", "build"}).
		WithExec([]string{"yarn", "run", "plugins:build-bundled"}).
		Directory("/src/public").
		WithoutDirectory("/src/public/app").
		WithoutDirectory("/src/public/test").
		WithoutDirectory("/src/public/testdata")

	return public
}

type YarnCacheOpts struct {
	// If HostDir is set, then that will be mounted to the .yarn/cache directory.
	HostDir *dagger.Directory

	// If HostDir is not set, then CacheVolume will be mounted in the container.
	CacheVolume *dagger.CacheVolume
}

type YarnInstallOpts struct {
	CacheOpts   *YarnCacheOpts
	Files       map[string]*dagger.File
	Directories map[string]*dagger.Directory
	NodeVersion string
}

func WithYarnCache(container *dagger.Container, opts *YarnCacheOpts) *dagger.Container {
	yarnCacheDir := "/yarn/cache"
	c := container.WithEnvVariable("YARN_CACHE_FOLDER", yarnCacheDir)
	if opts.HostDir != nil {
		return c.WithMountedDirectory(yarnCacheDir, opts.HostDir)
	}

	return c.WithMountedCache(yarnCacheDir, opts.CacheVolume)
}

func YarnInstall(ctx context.Context, d *dagger.Client, opts *YarnInstallOpts) error {
	container := NodeContainer(d, NodeImage(opts.NodeVersion)).
		WithWorkdir("/src")

	container = WithYarnCache(container, opts.CacheOpts)

	for path, file := range opts.Files {
		container = container.WithMountedFile(path, file)
	}

	for path, dir := range opts.Directories {
		container = container.WithMountedDirectory(path, dir)
	}

	container = container.
		WithExec([]string{"yarn", "install", "--immutable"})

	if e, err := container.ExitCode(ctx); err != nil {
		return fmt.Errorf("exit code '%d', error: %s", e, err.Error())
	}

	return nil
}
