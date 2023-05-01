package containers

import (
	"log"
	"path"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/executil"
)

const GoImage = "golang:1.20.1-alpine"

var GrafanaCommands = []string{
	"grafana",
	"grafana-server",
	"grafana-cli",
}

// BuilderPlatform returns the most optimal dagger.Platform for building the provided distribution.
// if 'platform' is 'amd64', then we can always use the 'amd64' platform.
// if 'platform' is 'arm64' and 'distro' is 'arm64', then we can use the 'arm64' platform. On Docker for Mac, this should be the most optimal for performance.
// if 'platform' is 'arm64' but 'distro' is 'amd64', we have to rely on buildkit's platform emulation. This will be very slow and really shouldn't be used.
// if 'platform' is 'arm64' but 'distro' is 'arm', then ... TBD?
func BuilderPlatform(distro executil.Distribution, platform dagger.Platform) dagger.Platform {
	_, arch := executil.OSAndArch(distro)
	switch arch {
	// If the distro's arch is arm64 then we use whatever platform is explicitly requested
	case "arm64":
		return platform
	// If the distro's arch is amd64, then we always use an amd64 platform
	case "amd64":
		return dagger.Platform("linux/amd64")
	default:
		return dagger.Platform("linux/amd64")
	}
}

// CompileBackendBuilder returns the container that is completely set up to build Grafana for the given distribution.
// * dir refers to the source tree of Grafana'; this could be a freshly cloned copy of Grafana.
// * buildinfo will be added as build flags to the 'go build' command.
func CompileBackendBuilder(d *dagger.Client, distro executil.Distribution, platform dagger.Platform, dir *dagger.Directory, buildinfo *BuildInfo) *dagger.Container {
	log.Println("Creating Grafana backend build container for", distro, "on platform", platform)

	// These options determine the CLI arguments and environment variables passed to the go build command.
	goBuildOptsFunc, ok := DistributionGoOpts[distro]
	if !ok {
		goBuildOptsFunc = DefaultBuildOpts
	}

	var (
		opts     = goBuildOptsFunc(distro, buildinfo)
		os, arch = executil.OSAndArch(distro)
		env      = executil.GoBuildEnv(opts)
	)

	builder := GolangContainer(d, BuilderPlatform(distro, platform), GoImage)

	// amd64 linux just needs a native go build using the golang container without setting CC and such.
	isAMD64Linux := os == "linux" && strings.Contains(arch, "amd64")
	if !isAMD64Linux {
		builder = ViceroyContainer(d, distro, ViceroyImage)
	}

	builder = builder.WithMountedDirectory("/src", dir).
		WithWorkdir("/src").
		WithExec([]string{"make", "gen-go"})

	// Fix: Avoid setting CC, GOOS, GOARCH when cross-compiling before `make gen-go` has been ran.
	if !isAMD64Linux {
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
		log.Printf("Building '%s' for %s", v, distro)
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
func CompileBackend(d *dagger.Client, distro executil.Distribution, platform dagger.Platform, dir *dagger.Directory, buildinfo *BuildInfo) *dagger.Directory {
	container := CompileBackendBuilder(d, distro, platform, dir, buildinfo)
	return BackendBinDir(container, distro)
}

func BackendBinDir(container *dagger.Container, distro executil.Distribution) *dagger.Directory {
	return container.Directory(path.Join("bin", string(distro)))
}
