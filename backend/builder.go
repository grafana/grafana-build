package backend

import (
	"errors"
	"fmt"
	"log/slog"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/golang"
)

// BuildOpts are general options that can change the way Grafana is compiled regardless of distribution.
type BuildOpts struct {
	Version           string
	ExperimentalFlags []string
	Tags              []string
	WireTag           string
	Static            bool
	Enterprise        bool
}

func distroOptsFunc(log *slog.Logger, distro Distribution) (DistroBuildOptsFunc, error) {
	if val, ok := DistributionGoOpts[distro]; ok {
		return DistroOptsLogger(log, val), nil
	}
	return nil, errors.New("unrecognized distirbution")
}

func WithGoEnv(log *slog.Logger, container *dagger.Container, distro Distribution, opts *BuildOpts) (*dagger.Container, error) {
	fn, err := distroOptsFunc(log, distro)
	if err != nil {
		return nil, err
	}
	bopts := fn(distro, opts.ExperimentalFlags, opts.Tags)

	return containers.WithEnv(container, GoBuildEnv(bopts)), nil
}

func WithViceroyEnv(log *slog.Logger, container *dagger.Container, distro Distribution, opts *BuildOpts) (*dagger.Container, error) {
	fn, err := distroOptsFunc(log, distro)
	if err != nil {
		return nil, err
	}
	bopts := fn(distro, opts.ExperimentalFlags, opts.Tags)

	return containers.WithEnv(container, ViceroyEnv(bopts)), nil
}

func ViceroyContainer(
	d *dagger.Client,
	log *slog.Logger,
	distro Distribution,
	goVersion string,
	viceroyVersion string,
	opts *BuildOpts,
) (*dagger.Container, error) {
	containerOpts := dagger.ContainerOpts{
		Platform: "linux/amd64",
	}

	// Instead of directly using the `arch` variable here to substitute in the GoURL, we have to be careful with the Go releases.
	// Supported releases (in the names):
	// * amd64
	// * armv6l
	// * arm64
	goURL := golang.DownloadURL(goVersion, "amd64")
	container := d.Container(containerOpts).From(fmt.Sprintf("rfratto/viceroy:%s", viceroyVersion))

	// Install Go manually, and install make, git, and curl from the package manager.
	container = container.WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-yq", "curl", "make", "git"}).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("curl -L %s | tar -C /usr/local -xzf -", goURL)}).
		WithEnvVariable("PATH", "/bin:/usr/bin:/usr/local/bin:/usr/local/go/bin:/usr/osxcross/bin")

	return WithViceroyEnv(log, container, distro, opts)
}

func GolangContainer(
	d *dagger.Client,
	log *slog.Logger,
	goVersion string,
	viceroyVersion string,
	platform dagger.Platform,
	distro Distribution,
	opts *BuildOpts,
) (*dagger.Container, error) {
	os, _ := OSAndArch(distro)
	// Only use viceroy for all darwin and only windows/amd64
	if os == "darwin" || distro == DistWindowsAMD64 {
		return ViceroyContainer(d, log, distro, goVersion, viceroyVersion, opts)
	}

	container := golang.Container(d, platform, goVersion).
		WithExec([]string{"apk", "add", "zig", "--repository=https://dl-cdn.alpinelinux.org/alpine/edge/testing"}).
		WithExec([]string{"apk", "add", "--update", "build-base", "alpine-sdk", "musl", "musl-dev"}).
		// Install the toolchain specifically for armv7 until we figure out why it's crashing w/ zig container = container.
		WithExec([]string{"mkdir", "/toolchain"}).
		WithExec([]string{"wget", "http://musl.cc/arm-linux-musleabihf-cross.tgz", "-P", "/toolchain"}).
		WithExec([]string{"tar", "-xvf", "/toolchain/arm-linux-musleabihf-cross.tgz", "-C", "/toolchain"})

	return WithGoEnv(log, container, distro, opts)
}

func withCue(c *dagger.Container, src *dagger.Directory) *dagger.Container {
	return c.
		WithDirectory("/src/cue.mod", src.Directory("cue.mod")).
		WithDirectory("/src/kinds", src.Directory("kinds")).
		WithDirectory("/src/packages/grafana-schema", src.Directory("packages/grafana-schema"), dagger.ContainerWithDirectoryOpts{
			Include: []string{"**/*.cue"},
		}).
		WithDirectory("/src/public/app/plugins", src.Directory("public/app/plugins"), dagger.ContainerWithDirectoryOpts{
			Include: []string{"**/*.cue", "**/plugin.json"},
		}).
		WithFile("/src/embed.go", src.File("embed.go"))
}

// Builder returns the container that is used to build the Grafana backend binaries.
// The build container:
// * Will be based on rfratto/viceroy for Darwin or Windows
// * Will be based on golang:x.y.z-alpine for all other ditsros
// * Will download & cache the downloaded Go modules
// * Will run `make gen-go` on the provided Grafana source
//   - With the linux/amd64 arch/os combination, regardless of what the requested distro is.
//
// * And will have all of the environment variables necessary to run `go build`.
func Builder(
	d *dagger.Client,
	log *slog.Logger,
	distro Distribution,
	opts *BuildOpts,
	platform dagger.Platform,
	src *dagger.Directory,
	goVersion string,
	viceroyVersion string,
) (*dagger.Container, error) {
	var (
		version  = opts.Version
		cacheKey = "go-mod-" + version
		cacheDir = golang.ModuleDir(d, platform, src, goVersion)
		cache    = d.CacheVolume(cacheKey)
	)

	// for some distros we use the golang official iamge. For others, we use viceroy.
	builder, err := GolangContainer(d, log, goVersion, viceroyVersion, platform, distro, opts)
	if err != nil {
		return nil, err
	}

	commitInfo := GetVCSInfo(d, src, version, opts.Enterprise)

	builder = withCue(builder, src).
		WithDirectory("/src",
			// to generate this list, run this from the root of grafana:  ls -a . | xargs -n1  -I '{}' echo 'WithoutFile("{}").'
			// and remove the files that are required for the build.
			src.
				WithoutDirectory(".bingo").
				WithoutDirectory(".changelog-archive").
				WithoutDirectory(".github").
				WithoutDirectory(".husky").
				WithoutDirectory(".vscode").
				WithoutDirectory("bin").
				WithoutDirectory("conf").
				WithoutDirectory("contribute").
				WithoutDirectory("cue.mod").
				WithoutDirectory("data").
				WithoutDirectory("devenv").
				WithoutDirectory("docs").
				WithoutDirectory("e2e").
				WithoutDirectory("emails").
				WithoutDirectory("grafana-mixin").
				WithoutDirectory("hack").
				WithoutDirectory("kinds").
				WithoutDirectory("local").
				WithoutDirectory("node_modules").
				WithoutDirectory("packages").
				WithoutDirectory("packaging").
				WithoutDirectory("plugins-bundled").
				WithoutDirectory("public").
				WithoutDirectory("scripts").
				WithoutDirectory("tools").
				WithoutFile(".betterer.results").
				WithoutFile(".betterer.results.json").
				WithoutFile(".betterer.ts").
				WithoutFile(".bra.toml").
				WithoutFile(".browserslistrc").
				WithoutFile("build.go").
				WithoutFile("CHANGELOG.md").
				WithoutFile("CODE_OF_CONDUCT.md").
				WithoutFile("CONTRIBUTING.md").
				WithoutFile("crowdin.yml").
				WithoutFile("cypress.config.js").
				WithoutFile("Dockerfile").
				WithoutFile(".dockerignore").
				WithoutFile(".drone.star").
				WithoutFile(".drone.yml").
				WithoutFile(".editorconfig").
				WithoutFile("embed.go").
				WithoutFile(".eslintignore").
				WithoutFile(".eslintrc").
				WithoutFile(".gitattributes").
				WithoutFile(".gitignore").
				WithoutFile(".golangci.toml").
				WithoutFile("GOVERNANCE.md").
				WithoutFile("HALL_OF_FAME.md").
				WithoutFile("jest.config.js").
				WithoutFile("latest.json").
				WithoutFile("lefthook.rc").
				WithoutFile("lefthook.yml").
				WithoutFile("lerna.json").
				WithoutFile(".levignore.js").
				WithoutFile("LICENSE").
				WithoutFile("LICENSING.md").
				WithoutFile("MAINTAINERS.md").
				WithoutFile("Makefile").
				WithoutFile("NOTICE.md").
				WithoutFile(".nvmrc").
				WithoutFile(".pa11yci.conf.js").
				WithoutFile(".pa11yci-pr.conf.js").
				WithoutFile("package.json").
				WithoutFile("playwright.config.ts").
				WithoutFile(".prettierignore").
				WithoutFile(".prettierrc.js").
				WithoutFile("README.md").
				WithoutFile("ROADMAP.md").
				WithoutFile("SECURITY.md").
				WithoutFile("stylelint.config.js").
				WithoutFile("SUPPORT.md").
				WithoutFile("tsconfig.json").
				WithoutFile("tsconfig.tsbuildinfo").
				WithoutFile("WORKFLOW.md").
				WithoutFile("yarn.lock").
				WithoutFile(".yarnrc.yml"),
		).
		WithFile("/src/pkg/server/wire_gen.go", Wire(d, src, platform, goVersion, opts.WireTag)).
		WithFile("/src/.buildinfo.commit", commitInfo.Commit).
		WithWorkdir("/src")

	if opts.Enterprise {
		builder = builder.WithFile("/src/.buildinfo.enterprise-commit", commitInfo.EnterpriseCommit)
	}

	builder = golang.WithCachedGoDependencies(
		builder,
		cacheDir, cache,
	)

	return builder, nil
}

func Wire(d *dagger.Client, src *dagger.Directory, platform dagger.Platform, goVersion string, wireTag string) *dagger.File {
	// withCue is only required during `make gen-go` in 9.5.x or older.
	return withCue(golang.Container(d, platform, goVersion), src).
		WithExec([]string{"apk", "add", "make"}).
		WithDirectory("/src",
			// to generate this list, run this from the root of grafana:  ls -a . | xargs -n1  -I '{}' echo 'WithoutFile("{}").'
			// and remove the files that are required for the build.
			src.
				WithoutDirectory(".changelog-archive").
				WithoutDirectory(".github").
				WithoutDirectory(".husky").
				WithoutDirectory(".vscode").
				WithoutDirectory("bin").
				WithoutDirectory("conf").
				WithoutDirectory("contribute").
				WithoutDirectory("cue.mod").
				WithoutDirectory("data").
				WithoutDirectory("devenv").
				WithoutDirectory("docs").
				WithoutDirectory("e2e").
				WithoutDirectory("emails").
				WithoutDirectory("grafana-mixin").
				WithoutDirectory("hack").
				WithoutDirectory("kinds").
				WithoutDirectory("local").
				WithoutDirectory("node_modules").
				WithoutDirectory("packages").
				WithoutDirectory("packaging").
				WithoutDirectory("plugins-bundled").
				WithoutDirectory("public").
				WithoutDirectory("scripts").
				WithoutDirectory("tools").
				WithoutFile(".betterer.results").
				WithoutFile(".betterer.results.json").
				WithoutFile(".betterer.ts").
				WithoutFile(".bra.toml").
				WithoutFile(".browserslistrc").
				WithoutFile("build.go").
				WithoutFile("CHANGELOG.md").
				WithoutFile("CODE_OF_CONDUCT.md").
				WithoutFile("CONTRIBUTING.md").
				WithoutFile("crowdin.yml").
				WithoutFile("cypress.config.js").
				WithoutFile("Dockerfile").
				WithoutFile(".dockerignore").
				WithoutFile(".drone.star").
				WithoutFile(".drone.yml").
				WithoutFile(".editorconfig").
				WithoutFile("embed.go").
				WithoutFile(".eslintignore").
				WithoutFile(".eslintrc").
				WithoutFile(".gitattributes").
				WithoutFile(".gitignore").
				WithoutFile(".golangci.toml").
				WithoutFile("GOVERNANCE.md").
				WithoutFile("HALL_OF_FAME.md").
				WithoutFile("jest.config.js").
				WithoutFile("latest.json").
				WithoutFile("lefthook.rc").
				WithoutFile("lefthook.yml").
				WithoutFile("lerna.json").
				WithoutFile(".levignore.js").
				WithoutFile("LICENSE").
				WithoutFile("LICENSING.md").
				WithoutFile("MAINTAINERS.md").
				WithoutFile("NOTICE.md").
				WithoutFile(".nvmrc").
				WithoutFile(".pa11yci.conf.js").
				WithoutFile(".pa11yci-pr.conf.js").
				WithoutFile("package.json").
				WithoutFile("playwright.config.ts").
				WithoutFile(".prettierignore").
				WithoutFile(".prettierrc.js").
				WithoutFile("README.md").
				WithoutFile("ROADMAP.md").
				WithoutFile("SECURITY.md").
				WithoutFile("stylelint.config.js").
				WithoutFile("SUPPORT.md").
				WithoutFile("tsconfig.json").
				WithoutFile("tsconfig.tsbuildinfo").
				WithoutFile("WORKFLOW.md").
				WithoutFile("yarn.lock").
				WithoutFile(".yarnrc.yml"),
		).
		WithWorkdir("/src").
		WithExec([]string{"make", "gen-go", fmt.Sprintf("WIRE_TAGS=%s", wireTag)}).
		File("/src/pkg/server/wire_gen.go")
}
