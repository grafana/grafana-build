package pipelines

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
	"github.com/grafana/grafana-build/versions"
)

// PackagedFiels are files that are included in the grafana tarball.
var PackagedFiles = []string{
	"Dockerfile",
	"LICENSE",
	"NOTICE.md",
	"README.md",
}

// PackagedFiels are files that are included in the grafana tarball.
var PackagedDirectories = []string{
	"conf/",
	"docs/sources/",
	"packaging/deb/",
	"packaging/rpm/",
	"packaging/docker/",
	"packaging/wrappers/",
	"plugins-bundled/",
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
		src         = opts.Source
		distros     = opts.Distributions
		version     = opts.Version
		versionOpts = versions.OptionsFor(version)
		buildID     = opts.BuildID
		edition     = opts.Edition
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
	if err := containers.YarnInstall(ctx, d, opts.Platform, &containers.YarnInstallOpts{
		NodeVersion: nodeVersion,
		Directories: map[string]*dagger.Directory{
			".yarn":           src.Directory(".yarn").WithoutDirectory("/src/.yarn/cache"),
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
		frontend    = containers.CompileFrontend(d, opts.Platform, src, cacheOpts, nodeVersion)
		npmPackages = containers.NPMPackages(d, opts.Platform, src, cacheOpts, version, nodeVersion)
		storybook   = containers.Storybook(d, opts.Platform, src, cacheOpts, version, nodeVersion)
	)

	root := fmt.Sprintf("grafana-%s", strings.TrimPrefix(version, "v"))
	packages := make(map[executil.Distribution]*dagger.File, len(backends))

	dirs := PackagedDirectories
	if versionOpts.Autocomplete.IsSet && versionOpts.Autocomplete.Value {
		dirs = append(dirs, "packaging/autocomplete")
	}

	files := map[string]*dagger.File{
		"VERSION": d.Container().From(containers.BusyboxImage).
			WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("echo \"%s\" > VERSION", opts.Version)}).File("VERSION"),
	}
	for _, v := range PackagedFiles {
		files[v] = src.File(v)
	}

	directories := map[string]*dagger.Directory{}
	for _, v := range dirs {
		directories[v] = src.Directory(v)
	}

	for k, backend := range backends {
		packageDirs := map[string]*dagger.Directory{
			"bin":           backend,
			"public":        frontend,
			"npm-artifacts": npmPackages,
			"storybok":      storybook,
		}

		for k, v := range directories {
			packageDirs[k] = v
		}

		opts := TarFileOpts{
			Version: version,
			BuildID: buildID,
			Edition: edition,
			Distro:  k,
		}
		name := TarFilename(opts)

		packager := d.Container().
			From(containers.BusyboxImage).
			WithEnvVariable("PACKAGE_NAME", name)

		packages[k] = containers.TargzFile(packager, &containers.TargzFileOpts{
			Root:        root,
			Directories: packageDirs,
			Files:       map[string]*dagger.File{},
		})
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
