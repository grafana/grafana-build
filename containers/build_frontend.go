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

func YarnInstall(d *dagger.Client, src *dagger.Directory, nodeVersion string) *dagger.Directory {
	// Get the node version from the 'src' directories '.nvmrc' file.
	return NodeContainer(d, NodeImage(nodeVersion)).
		WithFile("/src/package.json", src.File("package.json")).
		WithFile("/src/yarn.lock", src.File("yarn.lock")).
		WithFile("/src/.yarnrc.yml", src.File(".yarnrc.yml")).
		WithDirectory("/src/.yarn", src.Directory(".yarn")).
		WithDirectory("/src/packages", src.Directory("packages")).
		WithDirectory("/src/plugins-bundled", src.Directory("plugins-bundled")).
		WithWorkdir("/src").
		WithExec([]string{"yarn", "install", "--immutable"}).
		Directory("/src")
}

func CompileFrontend(d *dagger.Client, src *dagger.Directory, yarnInstall *dagger.Directory, nodeVersion string) *dagger.Directory {
	// Get the node version from the 'src' directories '.nvmrc' file.
	return NodeContainer(d, NodeImage(nodeVersion)).
		WithDirectory("/src", src).
		WithDirectory("/src/.yarn", yarnInstall.Directory("/.yarn")).
		WithDirectory("/src/node_modules", yarnInstall.Directory("/node_modules")).
		WithWorkdir("/src").
		WithExec([]string{"yarn", "install", "--immutable"}).
		WithExec([]string{"yarn", "run", "build"}).
		Directory("public/")
}
