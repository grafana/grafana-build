package pipelines

import (
	"context"
	"fmt"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
)

// PackagedPaths are paths that are included in the grafana tarball.
var PackagedPaths = []string{
	"bin/",
	"conf/",
	"LICENSE",
	"NOTICE.md",
	"plugins-bundled/",
	"public/",
	"README.md",
	"VERSION",
	"docs/sources/",
	"packaging/deb",
	"packaging/rpm",
	"packaging/wrappers",
	"packaging/autocomplete",
}

// The PackageOpts command requires all of the options to build Grafana, but supports a list of distros instead of just one.
// It also requires extra options for determining the package name.
// While this struct embeds GrafanaCompileOpts, it ignores the 'Distribution' field in favor of the 'Distributions' list.
type PackageOpts struct {
	*GrafanaCompileOpts
	Distributions []executil.Distribution

	BuildID string
	Edition string
}

// PackageFile builds and packages Grafana into a tar.gz for each dsitrbution and returns a map of the dagger file that holds each tarball, keyed by the distribution it corresponds to.
func PackageFiles(ctx context.Context, d *dagger.Client, opts PackageOpts) (map[executil.Distribution]*dagger.File, error) {
	var (
		src     = opts.Source
		distros = opts.Distributions
		version = opts.Version
		buildID = opts.BuildID
		edition = opts.Edition
	)
	backends, err := GrafanaBackendBuildDirectories(ctx, d, opts.GrafanaCompileOpts, distros)
	if err != nil {
		return nil, err
	}

	// Get the node version from .nvmrc...
	nodeVersion, err := containers.NodeVersion(d, src).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node version from source code: %w", err)
	}

	nodeCache := d.CacheVolume("yarn")
	frontend := containers.CompileFrontend(d, src, nodeCache, nodeVersion)
	if err != nil {
		return nil, err
	}

	packages := make(map[executil.Distribution]*dagger.File, len(backends))
	for k, backend := range backends {
		packager := d.Container().
			From(containers.BusyboxImage).
			WithMountedDirectory("/src", src).
			WithMountedDirectory("/src/bin", backend).
			WithMountedDirectory("/src/public", frontend).
			WithWorkdir("/src")

		opts := TarFileOpts{
			Version: version,
			BuildID: buildID,
			Edition: edition,
			Distro:  k,
		}

		name := TarFilename(opts)
		packager = packager.WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("echo \"%s\" > VERSION", opts.Version)}).
			WithExec(append([]string{"tar", "-czf", name}, PackagedPaths...))
		packages[k] = packager.File(name)
	}

	return packages, nil
}

// Package builds and packages Grafana into a tar.gz for each distribution provided.
func Package(ctx context.Context, d *dagger.Client, opts PackageOpts) error {
	packages, err := PackageFiles(ctx, d, opts)
	if err != nil {
		return err
	}

	for k, file := range packages {
		opts := TarFileOpts{
			Version: opts.Version,
			BuildID: opts.BuildID,
			Edition: opts.Edition,
			Distro:  k,
		}
		name := TarFilename(opts)
		if _, err := file.Export(ctx, name); err != nil {
			return err
		}
	}
	return nil
}
