package containers

import (
	"fmt"
	"log"
	"path"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/executil"
)

const GoImage = "golang:1.20.1-bullseye"

var GrafanaCommands = []string{
	"grafana",
	"grafana-server",
	"grafana-cli",
}

func CompileBackendBuilder(d *dagger.Client, distro executil.Distribution, dir *dagger.Directory, buildinfo *BuildInfo) *dagger.Container {
	log.Println("Creating Grafana backend build container for", distro)

	os, arch := executil.OSAndArch(distro)

	opts := &executil.GoBuildOpts{
		ExperimentalFlags: []string{},
		OS:                os,
		Arch:              arch,
		CGOEnabled:        true,
		TrimPath:          true,
		LDFlags: map[string][]string{
			"-w": nil,
			"-s": nil,
			"-X": {
				fmt.Sprintf("main.version=%s", buildinfo.Version),
				fmt.Sprintf("main.commit=%s", buildinfo.Commit),
				fmt.Sprintf("main.buildstamp=%d", buildinfo.Timestamp.Unix()),
				fmt.Sprintf("main.buildBranch=%s", buildinfo.Branch),
			},
			"-linkmode=external":  nil,
			"-extldflags=-static": nil,
		},
		Tags: []string{
			"netgo",
			"osusergo",
		},
	}

	if executil.IsWindows(distro) {
		opts.BuildMode = executil.BuildModeExe
	}

	var (
		env = executil.GoBuildEnv(opts)
	)

	platform := dagger.Platform("linux/amd64")
	if arch == "arm64" {
		platform = dagger.Platform("linux/arm64")
	}

	if arch == "arm" {
		platform = dagger.Platform("linux/arm")
	}

	builder := GolangContainer(d, platform, GoImage).
		WithMountedDirectory("/src", dir).
		WithWorkdir("/src").
		WithExec([]string{"make", "gen-go"})

	// If we're building for darwin...
	if os == "darwin" {
		if arch == "amd64" {
			builder = WithDarwinAMD64Toolchain(builder)
		}
		if arch == "arm64" {
			builder = WithDarwinARM64Toolchain(builder)
		}
	}

	for k, v := range env {
		builder = builder.WithEnvVariable(k, v)
	}

	for _, v := range GrafanaCommands {
		opts := opts
		opts.Main = path.Join("pkg", "cmd", v)
		opts.Output = path.Join("bin", string(distro), v)

		cmd := executil.GoBuildCmd(opts)
		log.Printf("Building '%s' with env: '%+v'", v, env)
		log.Printf("Building '%s' with command: '%+v'", v, cmd)
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
	return BackendBinDir(CompileBackendBuilder(d, distro, dir, buildinfo), distro)
}

func BackendBinDir(container *dagger.Container, distro executil.Distribution) *dagger.Directory {
	return container.Directory(path.Join("bin", string(distro)))
}
