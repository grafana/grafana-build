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

func TarFilename(args PipelineArgs) string {
	name := "grafana.tar.gz"
	if args.Enterprise {
		name = "grafana-enterprise.tar.gz"
	}

	return name
}

// PackageFile builds and packages Grafana into a tar.gz and returns the dagger file that holds the tarball.
func PackageFile(ctx context.Context, d *dagger.Client, args PipelineArgs) (*dagger.File, error) {
	var (
		src     = args.Grafana
		version = args.Context.String("version")
		distro  = executil.Distribution(args.Context.String("distro"))
	)

	backend, err := GrafanaBackendBuildDirectory(ctx, d, src, distro, version)
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

	packager := d.Container().
		From(containers.BusyboxImage).
		WithMountedDirectory("/src", args.Grafana).
		WithMountedDirectory("/src/bin", backend).
		WithMountedDirectory("/src/public", frontend).
		WithWorkdir("/src")

	for _, v := range plugins {
		packager = packager.WithMountedDirectory(path.Join("/src/plugins-bundled/internal", v.Name), v.Directory)
	}
	name := TarFilename(args)
	packager = packager.WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("echo \"%s\" > VERSION", version)}).
		WithExec(append([]string{"tar", "-czf", name}, PackagedPaths...))

	return packager.File(name), nil
}

// Package builds and packages Grafana into a tar.gz.
func Package(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	file, err := PackageFile(ctx, d, args)
	if err != nil {
		return err
	}

	name := TarFilename(args)
	if _, err := file.Export(ctx, name); err != nil {
		return err
	}
	return nil
}

// PublishPackage creates a package and publishes it to a Google Cloud Storage bucket.
func PublishPackage(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	targz, err := PackageFile(ctx, d, args)
	if err != nil {
		return err
	}

	var auth containers.GCPAuthenticator = &containers.GCPInheritedAuth{}
	if key := args.Context.String("key"); key != "" {
		auth = containers.NewGCPServiceAccount(key)
	}

	uploader, err := containers.GCSUploadFile(d, containers.GoogleCloudImage, auth, targz, args.Context.Path("destination"))
	if err != nil {
		return err
	}

	return containers.ExitError(ctx, uploader)
}
