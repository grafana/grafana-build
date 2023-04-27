package containers

import (
	"dagger.io/dagger"
)

// GrafanaContainer is the base Golang image with everything set up to build and test Grafana.
// Mostly that means that:
// * The grafana source code is mounted.
// * 'make' is installed.
// * the wire dependency graph has been generated (using 'make gen-go')
// * schemas have been generated (using 'make gen-cue')
func GrafanaContainer(d *dagger.Client, base string, grafana *dagger.Directory) *dagger.Container {
	return GolangContainer(d, base).
		WithMountedDirectory("/src", grafana).
		WithWorkdir("/src").
		WithEnvVariable("CODEGEN_VERIFY", "1").
		WithExec([]string{"make", "gen-go"}).
		WithExec([]string{"make", "gen-cue"})
}
