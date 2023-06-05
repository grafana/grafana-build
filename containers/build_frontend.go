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

func CompileFrontend(d *dagger.Client, src *dagger.Directory, nodeCache *dagger.CacheVolume, nodeVersion string) *dagger.Directory {
	// Get the node version from the 'src' directories '.nvmrc' file.
	return NodeContainer(d, NodeImage(nodeVersion)).
		WithDirectory("/src", src).
		WithMountedCache("/src/.yarn/cache", nodeCache).
		WithWorkdir("/src").
		WithExec([]string{"yarn", "install", "--immutable"}).
		WithExec([]string{"yarn", "run", "build"}).
		WithExec([]string{"yarn", "run", "plugins:build-bundled"}).
		Directory("public/")
}

type YarnInstallOpts struct {
	Cache       *dagger.CacheVolume
	Files       map[string]*dagger.File
	Directories map[string]*dagger.Directory
	NodeVersion string
}

func YarnInstall(ctx context.Context, d *dagger.Client, opts *YarnInstallOpts) error {
	container := NodeContainer(d, NodeImage(opts.NodeVersion)).
		WithWorkdir("/src")

	for path, file := range opts.Files {
		container = container.WithMountedFile(path, file)
	}

	for path, dir := range opts.Directories {
		container = container.WithMountedDirectory(path, dir)
	}

	container = container.WithMountedCache("/src/.yarn/cache", opts.Cache).
		WithExec([]string{"yarn", "install", "--immutable"})

	if e, err := container.ExitCode(ctx); err != nil {
		return fmt.Errorf("exit code '%d', error: %s", e, err.Error())
	}

	return nil
}
