package pipelines

import (
	"context"
	"fmt"
	"path"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
)

// PackagedPaths are paths that are included in the grafana tarball.
var PackagedPaths = []string{
	"Dockerfile",
	"LICENSE",
	"NOTICE.md",
	"README.md",
	"VERSION",
	"bin/",
	"conf/",
	"docs/sources/",
	"packaging/deb",
	"packaging/rpm",
	"packaging/docker",
	"packaging/wrappers",
	"packaging/autocomplete",
	"plugins-bundled/",
	"public/",
	"npm-artifacts/",
}

func PathsWithRoot(root string, paths []string) []string {
	p := make([]string, len(paths))
	for i, v := range paths {
		p[i] = path.Join(root, v)
	}

	return p
}

// The PackageOpts command requires all of the options to build Grafana, but supports a list of distros instead of just one.
// It also requires extra options for determining the package name.
// While this struct embeds GrafanaCompileOpts, it ignores the 'Distribution' field in favor of the 'Distributions' list.
type PackageOpts struct {
	*GrafanaCompileOpts
	Distributions []executil.Distribution

	BuildID         string
	Edition         string
	NodeCacheVolume *dagger.CacheVolume
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

	cacheOpts := &containers.YarnCacheOpts{
		CacheVolume: opts.NodeCacheVolume,
	}

	if opts.YarnCacheHostDir != "" {
		cacheOpts.HostDir = d.Host().Directory(opts.YarnCacheHostDir)
	}

	// install and cache the node modules
	if err := containers.YarnInstall(ctx, d, &containers.YarnInstallOpts{
		NodeVersion: nodeVersion,
		Directories: map[string]*dagger.Directory{
			".yarn":           src.Directory(".yarn").WithoutDirectory("cache"),
			"packages":        src.Directory("packages"),
			"plugins-bundled": src.Directory("plugins-bundled"),
		},
		Files: map[string]*dagger.File{
			"package.json": src.File("package.json"),
			"yarn.lock":    src.File("yarn.lock"),
			".yarnrc.yml":  src.File(".yarnrc.yml"),
		},
		CacheOpts: cacheOpts,
	}); err != nil {
		return nil, err
	}

	var (
		frontend    = containers.CompileFrontend(d, src, cacheOpts, nodeVersion)
		npmPackages = containers.NPMPackages(d, src, cacheOpts, version, nodeVersion)
	)

	name := "grafana"
	if edition != "" {
		name = fmt.Sprintf("%s-%s", name, edition)
	}

	root := fmt.Sprintf("%s-%s", name, version)

	packages := make(map[executil.Distribution]*dagger.File, len(backends))
	for k, backend := range backends {
		packager := d.Container().
			From(containers.BusyboxImage).
			WithMountedDirectory(path.Join("/src", root), src).
			WithMountedDirectory(path.Join("/src", root, "bin"), backend).
			WithMountedDirectory(path.Join("/src", root, "public"), frontend).
			WithMountedDirectory(path.Join("/src", root, "npm-artifacts"), npmPackages).
			WithWorkdir("/src")

		opts := TarFileOpts{
			Version: version,
			BuildID: buildID,
			Edition: edition,
			Distro:  k,
		}

		name := TarFilename(opts)
		packager = packager.WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("echo \"%s\" > %s", opts.Version, path.Join(root, "VERSION"))}).
			WithExec(append([]string{"tar", "-czf", name}, PathsWithRoot(root, PackagedPaths)...))
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
