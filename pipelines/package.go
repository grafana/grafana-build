package pipelines

import (
	"context"
	"fmt"
	"path"
	"strings"

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

// TarFileName returns a file name that matches this format: {grafana|grafana-enterprise}_{version}_{os}_{arch}_{build_number}.tar.gz
func TarFilename(version, buildID string, isEnterprise bool, distro executil.Distribution) string {
	name := "grafana"
	if isEnterprise {
		name = "grafana-enterprise"
	}
	var (
		// This should return something like "linux", "arm"
		os, arch = executil.OSAndArch(distro)
		// If applicable this will be set to something like "7" (for arm7)
		archv = executil.ArchVersion(distro)
	)
	if archv != "" {
		arch = strings.Join([]string{arch, archv}, "-")
	}

	p := []string{name, version, os, arch, buildID}

	return fmt.Sprintf("%s.tar.gz", strings.Join(p, "_"))
}

// PackageFile builds and packages Grafana into a tar.gz for each dsitrbution and returns a map of the dagger file that holds each tarball, keyed by the distribution it corresponds to.
func PackageFiles(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) (map[executil.Distribution]*dagger.File, error) {
	distros := executil.DistrosFromStringSlice(args.Context.StringSlice("distro"))

	backends, err := GrafanaBackendBuildDirectories(ctx, d, src, distros, args.Version)
	if err != nil {
		return nil, err
	}

	// Get the node version from .nvmrc...
	nodeVersion, err := containers.NodeVersion(d, src).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node version from source code: %w", err)
	}

	frontend, err := GrafanaFrontendBuildDirectory(ctx, d, src, nodeVersion)
	if err != nil {
		return nil, err
	}

	plugins, err := containers.BuildPlugins(ctx, d, src, "plugins-bundled/internal", nodeVersion)
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

		for _, v := range plugins {
			packager = packager.WithMountedDirectory(path.Join("/src/plugins-bundled/internal", v.Name), v.Directory)
		}
		name := TarFilename(args.Version, args.BuildID, args.BuildEnterprise, k)
		packager = packager.WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("echo \"%s\" > VERSION", args.Version)}).
			WithExec(append([]string{"tar", "-czf", name}, PackagedPaths...))
		packages[k] = packager.File(name)
	}

	return packages, nil
}

// Package builds and packages Grafana into a tar.gz for each distribution provided.
func Package(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) error {
	packages, err := PackageFiles(ctx, d, src, args)
	if err != nil {
		return err
	}

	for k, file := range packages {
		name := TarFilename(args.Version, args.BuildID, args.BuildEnterprise, k)
		if _, err := file.Export(ctx, name); err != nil {
			return err
		}
	}
	return nil
}
