package containers

import (
	"log"
	"path"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/executil"
)

const GoImage = "golang:1.20.1-alpine"

var GrafanaCommands = []string{
	"grafana",
	"grafana-server",
	"grafana-cli",
}

func CompileBackendBuilder(d *dagger.Client, distro executil.Distribution, dir *dagger.Directory, buildinfo *BuildInfo) *dagger.Container {
	log.Println("Creating Grafana backend build container for", distro)

	// These options determine the CLI arguments and environment variables passed to the go build command.
	goBuildOptsFunc, ok := DistributionGoOpts[distro]
	if !ok {
		goBuildOptsFunc = DefaultBuildOpts
	}

	var (
		opts     = goBuildOptsFunc(distro, buildinfo)
		os, arch = executil.OSAndArch(distro)
		env      = executil.GoBuildEnv(opts)
		platform = dagger.Platform("linux/amd64")
	)

	if arch == "arm64" {
		platform = dagger.Platform("linux/arm64")
	}

	if arch == "arm" {
		platform = dagger.Platform("linux/arm")
	}

	// For now, if we're not building for Linux, then we're going to be using rfratto/viceroy.
	builder := GolangContainer(d, platform, GoImage)
	if os != "linux" {
		builder = ViceroyContainer(d, distro, platform, ViceroyImage)
	}

	builder = builder.WithMountedDirectory("/src", dir).
		WithWorkdir("/src").
		WithExec([]string{"make", "gen-go"})

	// Fix: Avoid setting CC, GOOS, GOARCH when cross-compiling before `make gen-go` has been ran.
	if os != "linux" {
		builder = WithViceroyEnv(builder, opts)
	}

	for k, v := range env {
		builder = builder.WithEnvVariable(k, v)
	}

	for _, v := range GrafanaCommands {
		opts := opts
		opts.Main = path.Join("pkg", "cmd", v)
		opts.Output = path.Join("bin", string(distro), v)

		cmd := executil.GoBuildCmd(opts)
		log.Printf("Building '%s' on platform: '%+v'", v, platform)
		log.Printf("Building '%s' with env: '%+v'", v, env)
		log.Printf("Building '%s' with command: '%+v'", v, cmd)
		builder = builder.WithExec([]string{"env"})
		builder = builder.WithExec(cmd.Args)
	}

	return builder
}

type CompileConfig struct {
	// GrafanaPath is the relative or absolute path to the root of the Grafana source tree.
	// If empty, then GrafanaPath is assumed to be $PWD.
	GrafanaPath string

	// Version is injected into the binary at build-time using the ldflags compilation argument.
	Version string

	// Distributions are the different os/architecture combinations of binaries that are compiled
	Distributions []executil.Distribution
}

// CompileBackend returns a reference to a dagger directory that contains a usable Grafana binary from the cloned source code at 'grafanaPath'.
// The returned directory can be exported, which will cause the container to execute the build, or can be mounted into other containers.
func CompileBackend(d *dagger.Client, distro executil.Distribution, dir *dagger.Directory, buildinfo *BuildInfo) *dagger.Directory {
	container := CompileBackendBuilder(d, distro, dir, buildinfo)
	return BackendBinDir(container, distro)
}

func BackendBinDir(container *dagger.Container, distro executil.Distribution) *dagger.Directory {
	return container.Directory(path.Join("bin", string(distro)))
}
