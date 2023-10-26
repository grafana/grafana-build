package containers

import (
	"time"

	"dagger.io/dagger"
)

// Sha256 returns a dagger.File which contains the sha256 for the provided file.
func Sha256(d *dagger.Client, file *dagger.File) *dagger.File {
	return d.Container().From("busybox").
		WithEnvVariable("CACHE_DISABLE", time.Now().String()).
		WithFile("/src/file", file).
		WithExec([]string{"/bin/sh", "-c", "sha256sum /src/file | awk '{print $1}' > /src/file.sha256"}).
		File("/src/file.sha256")
}
