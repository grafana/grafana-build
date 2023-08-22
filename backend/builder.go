package backend

import "dagger.io/dagger"

// Builder returns the container that is used to build the Grafana backend binaries.
// This container needs to have:
// * zig, for cross-compilation
// * golang
// * musl
func Builder(d *dagger.Client, platform dagger.Platform) *dagger.Container {
	return nil
}
