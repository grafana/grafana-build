package containers

import (
	"path"

	"dagger.io/dagger"
)

type TargzFileOpts struct {
	// Root is the root folder that holds all of the packaged data.
	// It is common for targz packages to have a root folder.
	// This should equal something like `grafana-9.4.1`.
	Root string

	// A map of directory paths relative to the root, like 'bin', 'public', 'npm-artifacts'
	// to dagger directories.
	Directories map[string]*dagger.Directory
	Files       map[string]*dagger.File
}

func TargzFile(packager *dagger.Container, opts *TargzFileOpts) *dagger.File {
	root := opts.Root

	for k, v := range opts.Directories {
		packager = packager.
			WithMountedDirectory(path.Join("/src", root, k), v)
	}

	packager = packager.
		WithWorkdir("/src")

	paths := []string{}
	for k := range opts.Files {
		paths = append(paths, k)
	}

	for k := range opts.Directories {
		paths = append(paths, k)
	}

	packager = packager.WithExec(append([]string{"tar", "-czf", "/package.tar.gz"}, PathsWithRoot(root, paths)...))

	return packager.File("/package.tar.gz")
}

func PathsWithRoot(root string, paths []string) []string {
	p := make([]string, len(paths))
	for i, v := range paths {
		p[i] = path.Join(root, v)
	}

	return p
}
