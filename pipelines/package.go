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
	"bin/",
	"conf/",
	"LICENSE",
	"NOTICE.md",
	"plugins-bundled/",
	"public/",
	"README.md",
	"VERSION",
	"docs/sources/",
}

// Package builds and packages Grafana into a tar.gz.
func Package(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	var (
		src     = args.Grafana
		version = args.Context.String("version")
		distro  = executil.Distribution(args.Context.String("distro"))
	)

	backend, err := GrafanaBackendBuildDirectory(ctx, d, src, distro, version)
	if err != nil {
		return err
	}

	// Get the node version from .nvmrc...
	nodeVersion, err := containers.NodeVersion(d, src).Stdout(ctx)
	if err != nil {
		return fmt.Errorf("failed to get node version from source code: %w", err)
	}

	frontend, err := GrafanaFrontendBuildDirectory(ctx, d, src, nodeVersion)
	if err != nil {
		return err
	}

	plugins, err := containers.BuildPlugins(ctx, d, src, "plugins-bundled/internal", nodeVersion)
	if err != nil {
		return err
	}

	packager := d.Container().
		From(containers.BusyboxImage).
		WithMountedDirectory("/src", args.Grafana).
		WithMountedDirectory("/src/bin", backend).
		WithMountedDirectory("/src/public", frontend).
		WithWorkdir("/src")

	for _, v := range plugins {
		packager = packager.WithMountedDirectory(path.Join("/src/plugins-bundled/internal", v.Name), v.Directory)
	}

	packager = packager.WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("echo \"%s\" > VERSION", version)}).
		WithExec(append([]string{"tar", "-czf", "grafana.tar.gz"}, PackagedPaths...))

	if _, err := packager.File("grafana.tar.gz").Export(ctx, "grafana.tar.gz"); err != nil {
		return err
	}

	return nil
}
