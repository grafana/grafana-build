package pipelines

import (
	"context"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
)

// GrafanaCompileOpts are the options that must be supplied from the user to build Grafana.
type GrafanaCompileOpts struct {
	// Source is the source tree of Grafana, ideally retrieved from a git clone. This argument is required.
	Source *dagger.Directory

	// Distribution is the target distribution to compile for. This argument is required.
	Distribution executil.Distribution

	// Platform is the dagger platform to run the containers on. Unless there's a specific requirement, this should be left empty to match the current running docker platform.
	Platform dagger.Platform

	// Version is the semver version number to insert into the binary via build arguments.
	Version string

	// Env is an optional map of environment variables (keyed by literal variable name to value) to set in the build container(s).
	Env    map[string]string
	GoTags []string

	// Edition is just used for logging / visualization purposes
	Edition string
}

func (o *GrafanaCompileOpts) BuildInfo(ctx context.Context, d *dagger.Client) (*containers.BuildInfo, error) {
	return containers.GetBuildInfo(ctx, d, o.Source, o.Version)
}

func (o *GrafanaCompileOpts) BackendCompileOpts(ctx context.Context, d *dagger.Client) (*containers.CompileBackendOpts, error) {
	buildinfo, err := o.BuildInfo(ctx, d)
	if err != nil {
		return nil, err
	}

	return &containers.CompileBackendOpts{
		Source:       o.Source,
		Distribution: o.Distribution,
		Platform:     o.Platform,
		BuildInfo:    buildinfo,
	}, nil
}

// GrafanaBackendBuildDirectory returns a directory with the compiled backend binaries for the given distribution.
func GrafanaBackendBuildDirectory(ctx context.Context, d *dagger.Client, opts *GrafanaCompileOpts) (*dagger.Directory, error) {
	var (
		distro   = opts.Distribution
		version  = opts.Version
		platform = opts.Platform
		src      = opts.Source
	)
	if distro == "" {
		return nil, fmt.Errorf("not a valid distribution")
	}

	var (
		cacheKey = "go-mod-" + version
		cacheDir = containers.DownloadGolangDependencies(d, platform, src.File("go.mod"), src.File("go.sum"))
		cache    = d.CacheVolume(cacheKey)
	)

	backendCompileOpts, err := opts.BackendCompileOpts(ctx, d)
	if err != nil {
		return nil, err
	}

	container := containers.WithCachedGoDependencies(
		containers.CompileBackendBuilder(d, backendCompileOpts),
		cacheDir, cache,
	)

	return containers.BackendBinDir(container, distro), nil
}

// GrafanaBackendBuildDirectories builds multiple distributions and returns the directories for each one.
// The returned map of directories will be keyed by the distribution that the directory corresponds to.
func GrafanaBackendBuildDirectories(ctx context.Context, d *dagger.Client, opts *GrafanaCompileOpts, distros []executil.Distribution) (map[executil.Distribution]*dagger.Directory, error) {
	var (
		src      = opts.Source
		version  = opts.Version
		platform = opts.Platform
		env      = opts.Env
		tags     = opts.GoTags
	)
	if distros == nil {
		return nil, fmt.Errorf("distribution list can not be nil")
	}

	buildinfo, err := containers.GetBuildInfo(ctx, d, src, version)
	if err != nil {
		return nil, err
	}

	dirs := make(map[executil.Distribution]*dagger.Directory, len(distros))
	for _, distro := range distros {
		opts := &containers.CompileBackendOpts{
			Source:       src,
			Distribution: distro,
			Platform:     platform,
			BuildInfo:    buildinfo,
			Env:          env,
			GoTags:       tags,
		}

		dirs[distro] = containers.CompileBackend(d, opts)
	}

	return dirs, nil
}

// GrafanaBackendBuild builds all of the distributions in the '--distros' argument and places them in the 'bin' directory of the PWD.
func GrafanaBackendBuild(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) error {
	var (
		distroList = args.Context.StringSlice("distro")
		distros    = make([]executil.Distribution, len(distroList))
	)

	for i, v := range distroList {
		distros[i] = executil.Distribution(v)
	}

	dirs := make([]*dagger.Directory, len(distroList))
	for i, distro := range distros {
		opts := &GrafanaCompileOpts{
			Source:       src,
			Distribution: distro,
			Platform:     args.Platform,
			Version:      args.GrafanaOpts.Version,
			Env:          args.GrafanaOpts.Env,
			GoTags:       args.GrafanaOpts.GoTags,
		}
		container, err := GrafanaBackendBuildDirectory(ctx, d, opts)
		if err != nil {
			return err
		}

		dirs[i] = container
	}

	for i, v := range distros {
		var (
			dir    = dirs[i]
			output = filepath.Join("bin", string(v))
		)
		if _, err := dir.Export(ctx, output); err != nil {
			return err
		}
	}
	return nil
}
