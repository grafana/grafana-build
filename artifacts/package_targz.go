package artifacts

import (
	"context"
	"fmt"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/arguments"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/flags"
	"github.com/grafana/grafana-build/packages"
	"github.com/grafana/grafana-build/pipeline"
)

func TargzBuilder(ctx context.Context, o *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	state := o.State
	version, err := state.String(ctx, arguments.Version)
	if err != nil {
		return nil, err
	}

	container := o.Client.Container().
		From("alpine:3.18.4").
		WithExec([]string{"apk", "add", "--update", "tar"}).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("echo %s > VERSION", version)})

	return container, nil
}

func TargzBuild(ctx context.Context, o *pipeline.ArtifactBuildOpts) (*dagger.File, error) {
	state := o.State

	// The Grafana directory is used for other packaged data like Dockerfile, license.txt, etc.
	grafanaDir, err := state.Directory(ctx, arguments.GrafanaDirectory)
	if err != nil {
		return nil, err
	}

	backendDirFunc, err := o.Dependency(Backend)
	if err != nil {
		return nil, err
	}

  backendDir, err := backendDirFunc.Directory(ctx)
  if err != nil {
    return nil, err
  }

	frontendDirFunc, err := o.Dependency(Frontend)
	if err != nil {
		return nil, err
	}

  frontendDir, err := frontendDirFunc.Directory(ctx)
  if err != nil {
    return nil, err
  }

	files := map[string]*dagger.File{
    "VERSION": grafanaDir.File("VERSION"),
		"license.txt": grafanaDir.File("license.txt"),
	}

	directories := map[string]*dagger.Directory{
		"backend": backendDir,
		"public":  frontendDir,
	}
	root := ""

	return containers.TargzFile(
		o.Builder,
		&containers.TargzFileOpts{
			Root:        root,
			Files:       files,
			Directories: directories,
		},
	), nil
}

func PackageWithExtension(ext string) pipeline.FileNameFunc {
	return func(ctx context.Context, a pipeline.Artifact, state pipeline.StateHandler) (string, error) {
		return packages.ArtifactFilename(ctx, a, state, ext)
	}
}

var Tarball = pipeline.Artifact{
	Name: "targz",
	Type: pipeline.ArtifactTypeFile,
	Requires: []pipeline.Artifact{
		Backend,
		Frontend,
	},
	Flags: flags.StdPackageFlags(),
	Arguments: []pipeline.Argument{
		arguments.BuildID,
		arguments.Version,
	},
	FileNameFunc:  PackageWithExtension("tar.gz"),
	Builder:       TargzBuilder,
	BuildFileFunc: TargzBuild,
}
