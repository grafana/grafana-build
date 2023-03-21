package pipelines

import (
	"context"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
)

func GrafanaBackendBuildDirectory(ctx context.Context, d *dagger.Client, src *dagger.Directory, distro executil.Distribution, version string) (*dagger.Directory, error) {
	if distro == "" {
		return nil, fmt.Errorf("not a valid distribution")
	}

	var (
		cacheKey = "go-mod-" + version
		cacheDir = containers.DownloadGolangDependencies(d, src.File("go.mod"), src.File("go.sum"))
		cache    = d.CacheVolume(cacheKey)
	)

	buildinfo, err := containers.GetBuildInfo(ctx, d, src, version)
	if err != nil {
		return nil, err
	}

	container := containers.WithCachedGoDependencies(containers.CompileBackendBuilder(d, distro, src, buildinfo), cacheDir, cache)

	return containers.BackendBinDir(container, distro), nil
}

// GrafanaBackendBuildDirectories builds multiple distributions and returns the directories for each one.
// The returned map of directories will be keyed by the distribution that the directory corresponds to.
func GrafanaBackendBuildDirectories(ctx context.Context, d *dagger.Client, src *dagger.Directory, distros []executil.Distribution, version string) (map[executil.Distribution]*dagger.Directory, error) {
	if distros == nil {
		return nil, fmt.Errorf("distribution list can not be nil")
	}

	buildinfo, err := containers.GetBuildInfo(ctx, d, src, version)
	if err != nil {
		return nil, err
	}

	dirs := make(map[executil.Distribution]*dagger.Directory, len(distros))
	for _, distro := range distros {
		dirs[distro] = containers.CompileBackend(d, distro, src, buildinfo)
	}

	return dirs, nil
}

// GrafanaBackendBuild builds all of the distributions in the '--distros' argument and places them in the 'bin' directory of the PWD.
func GrafanaBackendBuild(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	var (
		version    = args.Context.String("version")
		distroList = args.Context.StringSlice("distro")
		distros    = make([]executil.Distribution, len(distroList))
	)

	for i, v := range distroList {
		distros[i] = executil.Distribution(v)
	}

	dirs := make([]*dagger.Directory, len(distroList))
	for i, distro := range distros {
		container, err := GrafanaBackendBuildDirectory(ctx, d, args.Grafana, distro, version)
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
