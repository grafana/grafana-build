package containers

import "dagger.io/dagger"

// ExtractedActive returns a directory that holds an extracted tar.gz
func ExtractedArchive(d *dagger.Client, f *dagger.File) *dagger.Directory {
	return d.Container().From("busybox").
		WithFile("/src/archive.tar.gz", f).
		WithExec([]string{"mkdir", "-p", "/src/archive"}).
		WithExec([]string{"tar", "--strip-components=1", "-xzf", "/src/archive.tar.gz", "-C", "/src/archive"}).
		Directory("/src/archive")
}
