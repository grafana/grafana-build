package containers

import (
	"fmt"
	"log"
	"path"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/executil"
	"github.com/grafana/grafana-build/versions"
)

const (
	GoImageAlpine = "golang:1.20.7-alpine"
)

var GrafanaCommands = []string{
	"grafana",
	"grafana-server",
	"grafana-cli",
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

func goBuildImage(distro executil.Distribution, opts *executil.GoBuildOpts) string {
	os, _ := executil.OSAndArch(distro)

	if os != "linux" {
		return ViceroyImage
	}

	return GoImageAlpine
}

// This function will return the platform equivalent to the distribution.
// However, if the distribution is a non-linux distribution, then it will always return linux/amd64.
// In the future it would probably be more suitable to also consider returning arm64 and using the viceroy:*-arm images.
func goBuildPlatform(distro executil.Distribution, platform dagger.Platform) dagger.Platform {
	os, _ := executil.OSAndArch(distro)
	// For non-linux containers we use use viceroy on linux/amd64.
	if os != "linux" {
		return dagger.Platform("linux/amd64")
	}

	// Otherwise we use the host's default platform.
	return platform
}

// CompileBackendBuilder returns the dagger container that will build the requested CompileBackendOpts.
// Goals:
//  0. Attempt to build Grafana using the platform given by the --platform argument.
//     The efficacy of this argument will depend on the docker buildkit capabilties, which can be checked with `docker buildx ls`.
//  1. If building for a different OS than Linux, then we use rfratto/vicery to accomplish that.
//     Almost all users' docker buildx capabilities will include linux/amd64, so this should at least be functional.
//     On Mac OS (especially using Apple Silicon chips), this will be incredibly slow. A 5 minute build on linux/amd64 could take 30+ minutes on darwin/arm64.
//  2. When building for `arm/v6` or `arm/v7` we want to build these exclusively on Alpine with musl.
//  3. When building anything statically we only want to build them on Alpine with musl.
func CompileBackendBuilder(d *dagger.Client, opts *CompileBackendOpts) *dagger.Container {
	var (
		distro    = opts.Distribution
		src       = opts.Source
		buildinfo = opts.BuildInfo
		platform  = goBuildPlatform(distro, opts.Platform)
	)

	// These options determine the CLI arguments and environment variables passed to the go build command.
	goBuildOptsFunc, ok := DistributionGoOpts[distro]
	if !ok {
		goBuildOptsFunc = DefaultBuildOpts
	}

	var (
		goBuildOpts = goBuildOptsFunc(distro, buildinfo)
		env         = executil.GoBuildEnv(goBuildOpts)
		image       = goBuildImage(distro, goBuildOpts)
	)

	log.Println("Creating Grafana backend build container for", distro, "on platform", platform)

	builder := GolangContainer(d, platform, image)

	if image == ViceroyImage {
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
	if image == ViceroyImage {
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

	// Add the user-provided opts.GoTags to the default list of tags.
	goBuildOpts.Tags = append(goBuildOpts.Tags, opts.GoTags...)

	for _, v := range commands {
		o := goBuildOpts
		o.Main = path.Join("pkg", "cmd", v)
		o.Output = path.Join("bin", string(distro), binaryName(v, distro))

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
