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
	log := o.Log

	log.Debug("Getting version from state for tar.gz builder...")
	version, err := state.String(ctx, arguments.Version)
	if err != nil {
		return nil, err
	}
	log.Debug("Got version from state for tar.gz builder", "version", version)

	container := o.Client.Container().
		From("alpine:3.18.4").
		WithExec([]string{"apk", "add", "--update", "tar"}).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("echo %s > VERSION", version)})

	return container, nil
}

func TargzBuild(ctx context.Context, o *pipeline.ArtifactBuildOpts) (*dagger.File, error) {
	state := o.ContainerOpts.State
	log := o.ContainerOpts.Log

	log.Debug("Getting grafana dir from state...")
	// The Grafana directory is used for other packaged data like Dockerfile, license.txt, etc.
	grafanaDir, err := state.Directory(ctx, arguments.GrafanaDirectory)
	if err != nil {
		return nil, err
	}
	log.Debug("Got grafana dir from state")

	log.Debug("Getting backend directory artifact...")
	backendDirFunc, err := o.Dependency(Backend)
	if err != nil {
		return nil, err
	}
	log.Debug("Got backend directory artifact")

	log.Debug("Getting backend directory from artifact...")
	backendDir, err := backendDirFunc.Directory(ctx, o.ContainerOpts)
	if err != nil {
		return nil, err
	}
	log.Debug("Got backend directory from artifact")

	log.Debug("Getting frontend directory from state...")
	frontendDirFunc, err := o.Dependency(Frontend)
	if err != nil {
		return nil, err
	}
	log.Debug("Got frontend directory from state")

	log.Debug("Getting frontend directory from artifact...")
	frontendDir, err := frontendDirFunc.Directory(ctx, o.ContainerOpts)
	if err != nil {
		return nil, err
	}
	log.Debug("Got frontend directory from artifact")

	files := map[string]*dagger.File{
		"VERSION":     grafanaDir.File("VERSION"),
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
