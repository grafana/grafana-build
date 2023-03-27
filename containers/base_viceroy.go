package containers

import (
	"fmt"
	"log"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/executil"
)

const (
	GoURL        = "https://go.dev/dl/go1.20.2.linux-%s.tar.gz"
	ViceroyImage = "rfratto/viceroy:v0.3.0"
)

// ViceroyContainer returns a dagger container with everything set up that is needed to build Grafana's Go backend
// with CGO using Viceroy, which makes setting up the C compiler toolchain easier.
func ViceroyContainer(d *dagger.Client, distro executil.Distribution, platform dagger.Platform, base string) *dagger.Container {
	opts := dagger.ContainerOpts{}
	if platform != "" {
		opts.Platform = platform
	}

	// Instead of directly using the `arch` variable here to substitute in the GoURL, we have to be careful with the Go releases.
	// Supported releases (in the names):
	// * amd64
	// * armv6l
	// * arm64
	goURL := fmt.Sprintf(GoURL, "amd64")
	if _, arch := executil.OSAndArch(distro); arch == "arm64" {
		goURL = fmt.Sprintf(GoURL, arch)
	}

	if _, arch := executil.OSAndArch(distro); arch == "arm" {
		goURL = fmt.Sprintf(GoURL, "armv6l")
	}

	container := d.Container(opts).From(base)

	// Install Go and other dependencies
	container = container.WithExec([]string{"apt-get", "update"})
	container = container.WithExec([]string{"apt-get", "install", "-yq", "curl", "make"})
	container = container.WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("curl -L %s | tar -C /usr/local -xzf -", goURL)})
	container = container.WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-yq", "git"})
	container = container.WithEnvVariable("PATH", "/bin:/usr/bin:/usr/local/bin:/usr/local/go/bin:/usr/osxcross/bin")

	return container
}

func WithViceroyEnv(container *dagger.Container, opts *executil.GoBuildOpts) *dagger.Container {
	log.Println("VICEROYOS", opts.OS)
	log.Println("VICEROYARCH", opts.Arch)
	container = container.
		WithEnvVariable("CC", "viceroycc").
		WithEnvVariable("VICEROYOS", opts.OS).
		WithEnvVariable("VICEROYARCH", opts.Arch)
	if opts.GoARM != "" {
		container = container.WithEnvVariable("VICEROYARM", string(opts.GoARM))
	}

	return container
}
