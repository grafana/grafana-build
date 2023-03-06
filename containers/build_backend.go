package containers

import (
	"fmt"
	"log"
	"path"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/executil"
)

const GoImage = "golang:1.20.1"

var DefaultDistros = []executil.Distribution{executil.DistDarwinAMD64, executil.DistDarwinARM64, executil.DistLinuxAMD64, executil.DistLinuxARM, executil.DistLinuxARM64, executil.DistWindowsAMD64}

func compileBackendBuilder(d *dagger.Client, distro executil.Distribution, dir *dagger.Directory, buildinfo *BuildInfo) *dagger.Container {
	opts := &executil.GoBuildOpts{
		Main:              path.Join("pkg", "cmd", "grafana"),
		Output:            path.Join("bin", string(distro), "grafana"),
		ExperimentalFlags: []string{},
		Distribution:      distro,
		CGOEnabled:        false,
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
		cmd = executil.GoBuildCmd(opts)
		env = executil.GoBuildEnv(opts)
	)

	log.Println("Building backend with env", env)
	log.Println("Building backend with command", cmd)
	builder := GolangContainer(d, GoImage).
		WithMountedDirectory("/src", dir).
		WithWorkdir("/src").
		WithExec([]string{"make", "gen-go"}).
		WithExec(cmd.Args)

	for k, v := range env {
		builder = builder.WithEnvVariable(k, v)
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
	return compileBackendBuilder(d, distro, dir, buildinfo).Directory(path.Join("bin", string(distro)))
}
