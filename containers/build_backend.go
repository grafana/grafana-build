package containers

import (
	"fmt"
	"log"
	"path"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/executil"
	"github.com/grafana/grafana-build/versions"
)

const (
	GoImageAlpine = "golang:1.20.1-alpine"
	GoImageDeb    = "golang:1.20.1"
)

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

func binaryName(command string, distro executil.Distribution) string {
	os, _ := executil.OSAndArch(distro)
	if os == "windows" {
		return command + ".exe"
	}

	return command
}

// CompileBackendOpts is similar to pipelines.CompileGrafanaOpts, but with more options specific to the backend compilation.
// CompileBackendOpts defines the options that are required to build the Grafana backend for a single distribution.
type CompileBackendOpts struct {
	Distribution executil.Distribution
	Platform     dagger.Platform
	Source       *dagger.Directory
	BuildInfo    *BuildInfo
	Env          map[string]string
	GoTags       []string

	CombinedExecutables bool
}

// CompileBackendBuilder returns the container that is completely set up to build Grafana for the given distribution.
// * dir refers to the source tree of Grafana'; this could be a freshly cloned copy of Grafana.
// * buildinfo will be added as build flags to the 'go build' command.
func CompileBackendBuilder(d *dagger.Client, opts *CompileBackendOpts) *dagger.Container {
	var (
		distro    = opts.Distribution
		platform  = opts.Platform
		src       = opts.Source
		buildinfo = opts.BuildInfo
	)
	log.Println("Creating Grafana backend build container for", distro, "on platform", platform)

	// These options determine the CLI arguments and environment variables passed to the go build command.
	goBuildOptsFunc, ok := DistributionGoOpts[distro]
	if !ok {
		goBuildOptsFunc = DefaultBuildOpts
	}

	var (
		goBuildOpts = goBuildOptsFunc(distro, buildinfo)
		env         = executil.GoBuildEnv(goBuildOpts)
	)

	goImage := GoImageAlpine
	switch goBuildOpts.LibC {
	case executil.Musl:
		goImage = GoImageAlpine
	case executil.GLibC:
		goImage = GoImageDeb
	}

	builder := GolangContainer(d, BuilderPlatform(distro, platform), goImage)

	// We are doing a "cross build" or cross compilation if the requested platform does not match the platform we're building for.
	isCrossBuild := !strings.Contains(string(distro), string(platform))
	if isCrossBuild {
		builder = ViceroyContainer(d, distro, ViceroyImage)
	}

	genGoArgs := []string{"make", "gen-go"}

	// This would give something like "make gen-go WIRE_TAG=minimal" if an env 'WIRE_TAG' was set to 'minimal'.
	// This is a workaround to our current Makefile not allowing overriding via environment variables.
	for k, v := range opts.Env {
		genGoArgs = append(genGoArgs, fmt.Sprintf("%s=%s", k, v))
	}

	// Ensure that we set the user provided `opts.Env` when running make gen-go.
	builder = builder.WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithExec(genGoArgs)

	// Adding env after the previous 'WithExec' ensures that when cross-building we still use the selected platform for 'make gen-go'.
	if isCrossBuild {
		builder = WithViceroyEnv(builder, goBuildOpts)
	}

	// Add the Go environment variables first, and then we can add the user-provided environment variables.
	// That way the user-provided ones can override the default ones.
	builder = WithEnv(builder, env)
	// TODO: we are doing this twice; once before make gen-go, and then again after. Would be nice if we only had to do this once.
	builder = WithEnv(builder, opts.Env)

	vopts := versions.OptionsFor(buildinfo.Version)
	commands := GrafanaCommands

	// If this version didn't support the combined executables, then only build grafana-server and grafana-cli
	if vopts.CombinedExecutable.IsSet && !vopts.CombinedExecutable.Value {
		commands = []string{
			"grafana-server",
			"grafana-cli",
		}
	}

	for _, v := range commands {
		o := goBuildOpts
		o.Main = path.Join("pkg", "cmd", v)
		o.Output = path.Join("bin", string(distro), binaryName(v, distro))
		o.Tags = opts.GoTags

		cmd := executil.GoBuildCmd(o)
		log.Printf("Building '%s' for %s", v, distro)
		log.Printf("Building '%s' with env: '%+v'", v, env)
		log.Printf("Building '%s' with command: '%+v'", v, cmd)
		builder = builder.WithExec(cmd.Args)
	}

	return builder
}

// CompileBackend returns a reference to a dagger directory that contains a usable Grafana binary from the cloned source code at 'grafanaPath'.
// The returned directory can be exported, which will cause the container to execute the build, or can be mounted into other containers.
func CompileBackend(d *dagger.Client, opts *CompileBackendOpts) *dagger.Directory {
	container := CompileBackendBuilder(d, opts)
	return BackendBinDir(container, opts.Distribution)
}

func BackendBinDir(container *dagger.Container, distro executil.Distribution) *dagger.Directory {
	return container.Directory(path.Join("bin", string(distro)))
}
