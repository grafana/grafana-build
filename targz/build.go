package targz

import (
	"path"

	"dagger.io/dagger"
)

type MappedDirectory struct {
	Path      string
	Directory *dagger.Directory
}

type MappedFile struct {
	Path string
	File *dagger.File
}

type Opts struct {
	// Root is the root folder that holds all of the packaged data.
	// It is common for targz packages to have a root folder.
	// This should equal something like `grafana-9.4.1`.
	Root string

	// A map of directory paths relative to the root, like 'bin', 'public', 'npm-artifacts'
	// to dagger directories.
	Directories []MappedDirectory
	Files       []MappedFile
}

func Build(packager *dagger.Container, opts *Opts) *dagger.File {
	root := opts.Root

	packager = packager.
		WithWorkdir("/src")

	paths := []string{}
	for _, v := range opts.Files {
		path := path.Join(root, v.Path)
		packager = packager.WithMountedFile(path, v.File)
		paths = append(paths, path)
	}

	for _, v := range opts.Directories {
		path := path.Join(root, v.Path)
		packager = packager.WithMountedDirectory(path, v.Directory)
		paths = append(paths, path)
	}

	packager = packager.WithExec(append([]string{"tar", "-czf", "/package.tar.gz"}, paths...))

	return packager.File("/package.tar.gz")
}
