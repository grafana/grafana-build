package containers

import (
	"dagger.io/dagger"
)

// NodeVersionContainer returns a container whose `stdout` will return the node version from the '.nvmrc' file in the directory 'src'.
func NodeVersion(d *dagger.Client, src *dagger.Directory) *dagger.Container {
	return d.Container().From("alpine:3.17").
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"cat", ".nvmrc"})
}

func CompileFrontend(d *dagger.Client, src *dagger.Directory, nodeVersion string) *dagger.Directory {
	// Get the node version from the 'src' directories '.nvmrc' file.
	return NodeContainer(d, NodeImage(nodeVersion)).
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"yarn", "install", "--immutable"}).
		WithExec([]string{"yarn", "run", "build"}).
		Directory("public")
}
