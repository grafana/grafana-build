package pipelines

import (
	"context"
	"fmt"
	"path"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
	"github.com/grafana/grafana-build/versions"
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
	if err := containers.YarnInstall(ctx, d, &containers.YarnInstallOpts{
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
		frontend    = containers.CompileFrontend(d, src, cacheOpts, nodeVersion)
		npmPackages = containers.NPMPackages(d, src, cacheOpts, version, nodeVersion)
	)

	name := "grafana"
	if edition != "" {
		name = fmt.Sprintf("%s-%s", name, edition)
	}

	root := fmt.Sprintf("%s-%s", name, version)

	packages := make(map[executil.Distribution]*dagger.File, len(backends))

	paths := PackagedPaths
	if versionOpts.Autocomplete.IsSet && versionOpts.Autocomplete.Value {
		paths = append(paths, "packaging/autocomplete")
	}

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
			WithExec(append([]string{"tar", "-czf", name}, PathsWithRoot(root, paths)...))
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

func getTarFileOpts(genOpts ArtifactGeneratorOptions) TarFileOpts {
	return TarFileOpts{
		Version: genOpts.PipelineArgs.GrafanaOpts.Version,
		BuildID: genOpts.PipelineArgs.GrafanaOpts.BuildID,
		Edition: genOpts.PipelineArgs.PackageOpts.Edition,
		Distro:  genOpts.Distribution,
	}
}

func GenerateTarballDirectory(ctx context.Context, d *dagger.Client, src *dagger.Directory, genOpts ArtifactGeneratorOptions, mounts map[string]*dagger.Directory) (*dagger.Directory, error) {
	args := genOpts.PipelineArgs
	root := "grafana"
	name := TarFilename(getTarFileOpts(genOpts))
	version := genOpts.PipelineArgs.GrafanaOpts.Version
	versionOpts := versions.OptionsFor(version)

	nodeVersion, err := containers.NodeVersion(d, src).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node version from source code: %w", err)
	}

	cacheOpts := &containers.YarnCacheOpts{
		CacheVolume: genOpts.NodeCacheVolume,
	}

	if genOpts.PipelineArgs.GrafanaOpts.YarnCacheHostDir != "" {
		cacheOpts.HostDir = d.Host().Directory(genOpts.PipelineArgs.GrafanaOpts.YarnCacheHostDir)
	}

	// install and cache the node modules
	if err := containers.YarnInstall(ctx, d, &containers.YarnInstallOpts{
		NodeVersion: nodeVersion,
		Directories: map[string]*dagger.Directory{
			".yarn":           src.Directory(".yarn").WithoutDirectory(".cache"),
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

	frontend := containers.CompileFrontend(d, src, cacheOpts, nodeVersion)
	npmPackages := containers.NPMPackages(d, src, cacheOpts, version, nodeVersion)

	paths := PackagedPaths
	if versionOpts.Autocomplete.IsSet && versionOpts.Autocomplete.Value {
		paths = append(paths, "packaging/autocomplete")
	}

	packager := d.Container().
		From(containers.BusyboxImage).
		WithExec([]string{"mkdir", "/dist"}).
		WithMountedDirectory("/src/grafana", src).
		WithMountedDirectory("/src/grafana/public", frontend).
		WithMountedDirectory(path.Join("/src", root, "npm-artifacts"), npmPackages).
		WithWorkdir("/src")

	for mountPoint, dir := range mounts {
		packager = packager.WithMountedDirectory(mountPoint, dir)
	}

	packager = packager.
		// TODO: Use container.WithNewFile here
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("echo \"%s\" > %s", args.GrafanaOpts.Version, path.Join(root, "VERSION"))}).
		WithExec(append([]string{"tar", "-czf", filepath.Join("/dist", name)}, PathsWithRoot(root, paths)...))
	return packager.Directory("/dist"), nil
}
