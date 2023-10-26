package containers

import "dagger.io/dagger"

func WithFiles(c *dagger.Container, files map[string]*dagger.File) *dagger.Container {
	container := c
	for path, file := range files {
		container = container.WithFile(path, file)
	}

	return container
}

func WithDirectories(c *dagger.Container, dirs map[string]*dagger.Directory) *dagger.Container {
	container := c
	for path, dir := range dirs {
		container = container.WithMountedDirectory(path, dir)
	}

	return container
}
