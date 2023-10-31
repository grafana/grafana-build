package targz

import (
	"path"

	"dagger.io/dagger"
)

type Opts struct {
	// Root is the root folder that holds all of the packaged data.
	// It is common for targz packages to have a root folder.
	// This should equal something like `grafana-9.4.1`.
	Root string

	// A map of directory paths relative to the root, like 'bin', 'public', 'npm-artifacts'
	// to dagger directories.
	Directories map[string]*dagger.Directory
	Files       map[string]*dagger.File
}

func Build(packager *dagger.Container, opts *Opts) *dagger.File {
	root := opts.Root

	for k, v := range opts.Directories {
		packager = packager.
			WithMountedDirectory(path.Join("/src", root, k), v)
	}

	packager = packager.
		WithWorkdir("/src")

	paths := []string{}
	for k, v := range opts.Files {
		path := path.Join(root, k)
		packager = packager.WithMountedFile(path, v)
		paths = append(paths, path)
	}

	for k, v := range opts.Directories {
		path := path.Join(root, k)
		packager = packager.WithMountedDirectory(path, v)
		paths = append(paths, path)
	}

	packager = packager.WithExec(append([]string{"tar", "-czf", "/package.tar.gz"}, paths...))

	return packager.File("/package.tar.gz")
}
