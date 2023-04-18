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
func GrafanaContainer(d *dagger.Client, platform dagger.Platform, base string, grafana *dagger.Directory) *dagger.Container {
	return GolangContainer(d, platform, base).
		WithMountedDirectory("/src", grafana).
		WithWorkdir("/src").
		WithEnvVariable("CODEGEN_VERIFY", "1").
		WithExec([]string{"make", "gen-go"})
}

// GrafanaContainerWithMounts returns the base golang iamge with everything needed for building / testing Grafana, but with only certain directories mounted.
// Using this function allows us to take advantage of caching a little bit better by only mounting the directories that are needed for a specific operation.
// Instead of mounting the entire Grafana directory at '/src' for `make gen-go` and `make gen-cue` like in 'GrafanaContainer(...)', we only mount the files and directories needed for these initialization operations as well.
func GrafanaContainerWithMounts(d *dagger.Client, platform dagger.Platform, base string, grafana *dagger.Directory, mounts map[string]*dagger.Directory) *dagger.Container {
	return GolangContainer(d, platform, base).
		WithMountedFile("/src/Makefile", grafana.File("Makefile")).
		WithMountedFile("/src/go.mod", grafana.File("go.mod")).
		WithMountedFile("/src/go.sum", grafana.File("go.sum")).
		WithMountedFile("/src/embed.go", grafana.File("go.sum")).
		WithMountedDirectory("/src/.bingo", grafana.Directory(".bingo")).
		WithMountedDirectory("/src/pkg", grafana.Directory("pkg")).
		WithMountedDirectory("/src/kinds", grafana.Directory("kinds")).
		WithMountedDirectory("/src/packages/grafana-schema/src/common", grafana.Directory("packages/grafana-schema/src/common")).
		WithMountedDirectory("/src/public/app/plugins", grafana.Directory("public/app/plugins")).
		WithWorkdir("/src").
		WithEnvVariable("CODEGEN_VERIFY", "1").
		WithExec([]string{"make", "gen-go"})
}
