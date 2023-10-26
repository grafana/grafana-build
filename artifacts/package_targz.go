package artifacts

import (
	"context"
	"fmt"
	"log/slog"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/arguments"
	"github.com/grafana/grafana-build/backend"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/flags"
	"github.com/grafana/grafana-build/packages"
	"github.com/grafana/grafana-build/pipeline"
)

var (
	TargzArguments = []pipeline.Argument{
		// Tarballs need the Build ID and version for naming the package properly.
		arguments.BuildID,
		arguments.Version,

		// The grafanadirectory has contents like the LICENSE.txt and such that need to be included in the package
		arguments.GrafanaDirectory,

		// The go version used to build the backend
		arguments.GoVersion,
		arguments.ViceroyVersion,
		arguments.YarnCacheDirectory,
	}
	TargzFlags = flags.JoinFlags(
		flags.StdPackageFlags(),
	)
)

var TargzInitializer = Initializer{
	InitializerFunc: NewTarballFromString,
	Arguments:       TargzArguments,
}

type Tarball struct {
	Distribution backend.Distribution
	Name         packages.Name
	BuildID      string
	Version      string

	Grafana *dagger.Directory

	// Dependent artifacts
	Backend        *pipeline.Artifact
	Frontend       *pipeline.Artifact
	NPMPackages    *pipeline.Artifact
	BundledPlugins *pipeline.Artifact
	Storybook      *pipeline.Artifact
}

func NewTarballFromString(ctx context.Context, log *slog.Logger, artifact string, state pipeline.StateHandler) (*pipeline.Artifact, error) {
	goVersion, err := state.String(ctx, arguments.GoVersion)
	if err != nil {
		return nil, err
	}
	viceroyVersion, err := state.String(ctx, arguments.ViceroyVersion)
	if err != nil {
		return nil, err
	}

	// 1. Figure out the options that were provided as part of the artifact string.
	//    For example, `linux/amd64:grafana`.
	options, err := pipeline.ParseFlags(artifact, TargzFlags)
	if err != nil {
		return nil, err
	}
	static, err := options.Bool(flags.Static)
	if err != nil {
		return nil, err
	}

	wireTag, err := options.String(flags.WireTag)
	if err != nil {
		return nil, err
	}

	experiments, err := options.StringSlice(flags.GoExperiments)
	if err != nil {
		return nil, err
	}

	yarnCache, err := state.CacheVolume(ctx, arguments.YarnCacheDirectory)
	if err != nil {
		return nil, err
	}

	p, err := GetPackageDetails(ctx, options, state)
	if err != nil {
		return nil, err
	}
	log.Info("Initializing tar.gz artifact with options", "name", p.Name, "build ID", p.BuildID, "version", p.Version, "distro", p.Distribution, "static", static, "enterprise", p.Enterprise)

	src, err := GrafanaDir(ctx, state, p.Enterprise)
	if err != nil {
		return nil, err
	}
	return NewTarball(ctx, log, artifact, p.Distribution, p.Name, p.Version, p.BuildID, src, yarnCache, static, wireTag, goVersion, viceroyVersion, experiments)
}

// NewTarball returns a properly initialized Tarball artifact.
// There are a lot of options that can affect how a tarball is built; most of which define different ways for the backend to be built.
func NewTarball(
	ctx context.Context,
	log *slog.Logger,
	artifact string,
	distro backend.Distribution,
	name packages.Name,
	version string,
	buildID string,
	src *dagger.Directory,
	cache *dagger.CacheVolume,
	static bool,
	wireTag string,
	goVersion string,
	viceroyVersion string,
	experiments []string,
) (*pipeline.Artifact, error) {
	backendArtifact, err := NewBackend(ctx, log, artifact, &NewBackendOpts{
		Name:           name,
		Version:        version,
		Distribution:   distro,
		Src:            src,
		Static:         static,
		WireTag:        wireTag,
		GoVersion:      goVersion,
		ViceroyVersion: viceroyVersion,
		Experiments:    experiments,
	})
	if err != nil {
		return nil, err
	}
	frontendArtifact, err := NewFrontend(ctx, log, artifact, name, src, cache)
	if err != nil {
		return nil, err
	}

	bundledPluginsArtifact, err := NewBundledPlugins(ctx, log, artifact, src, version, cache)
	if err != nil {
		return nil, err
	}

	npmArtifact, err := NewNPMPackages(ctx, log, artifact, src, version, cache)
	if err != nil {
		return nil, err
	}

	storybookArtifact, err := NewStorybook(ctx, log, artifact, src, version, cache)
	if err != nil {
		return nil, err
	}
	tarball := &Tarball{
		Name:         name,
		Distribution: distro,
		Version:      version,
		BuildID:      buildID,
		Grafana:      src,

		Backend:        backendArtifact,
		Frontend:       frontendArtifact,
		NPMPackages:    npmArtifact,
		BundledPlugins: bundledPluginsArtifact,
		Storybook:      storybookArtifact,
	}

	return pipeline.ArtifactWithLogging(ctx, log, &pipeline.Artifact{
		ArtifactString: artifact,
		Handler:        tarball,
		Type:           pipeline.ArtifactTypeFile,
		Flags:          TargzFlags,
	})
}

func (t *Tarball) Builder(ctx context.Context, opts *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	version := t.Version

	container := opts.Client.Container().
		From("alpine:3.18.4").
		WithExec([]string{"apk", "add", "--update", "tar"}).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("echo %s > VERSION", version)})

	return container, nil

}

func (t *Tarball) BuildFile(ctx context.Context, b *dagger.Container, opts *pipeline.ArtifactContainerOpts) (*dagger.File, error) {
	var (
		state = opts.State
		log   = opts.Log
	)

	log.Debug("Getting grafana dir from state...")
	// The Grafana directory is used for other packaged data like Dockerfile, license.txt, etc.
	grafanaDir := t.Grafana

	backendDir, err := opts.Store.Directory(ctx, t.Backend)
	if err != nil {
		return nil, err
	}

	frontendDir, err := opts.Store.Directory(ctx, t.Frontend)
	if err != nil {
		return nil, err
	}

	npmDir, err := opts.Store.Directory(ctx, t.NPMPackages)
	if err != nil {
		return nil, err
	}

	storybookDir, err := opts.Store.Directory(ctx, t.Storybook)
	if err != nil {
		return nil, err
	}

	pluginsDir, err := opts.Store.Directory(ctx, t.BundledPlugins)
	if err != nil {
		return nil, err
	}

	version, err := state.String(ctx, arguments.Version)
	if err != nil {
		return nil, err
	}

	files := map[string]*dagger.File{
		"VERSION":    b.File("VERSION"),
		"LICENSE":    grafanaDir.File("LICENSE"),
		"NOTICE.md":  grafanaDir.File("NOTICE.md"),
		"README.md":  grafanaDir.File("README.md"),
		"Dockerfile": grafanaDir.File("Dockerfile"),
	}

	directories := map[string]*dagger.Directory{
		"conf":               grafanaDir.Directory("conf"),
		"docs/sources":       grafanaDir.Directory("docs/sources"),
		"packaging/deb":      grafanaDir.Directory("packaging/deb"),
		"packaging/rpm":      grafanaDir.Directory("packaging/rpm"),
		"packaging/docker":   grafanaDir.Directory("packaging/docker"),
		"packaging/wrappers": grafanaDir.Directory("packaging/wrappers"),
		"bin":                backendDir,
		"public":             frontendDir,
		"npm-artifacts":      npmDir,
		"storybook":          storybookDir,
		"plugins-bundled":    pluginsDir,
	}

	root := fmt.Sprintf("grafana-%s", version)

	return containers.TargzFile(
		b,
		&containers.TargzFileOpts{
			Root:        root,
			Files:       files,
			Directories: directories,
		},
	), nil
}

func (t *Tarball) BuildDir(ctx context.Context, builder *dagger.Container, opts *pipeline.ArtifactContainerOpts) (*dagger.Directory, error) {
	panic("not implemented") // TODO: Implement
}

func (t *Tarball) Publisher(ctx context.Context, opts *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	panic("not implemented") // TODO: Implement
}

func (t *Tarball) PublishFile(ctx context.Context, opts *pipeline.ArtifactPublishFileOpts) error {
	panic("not implemented") // TODO: Implement
}

func (t *Tarball) PublisDir(ctx context.Context, opts *pipeline.ArtifactPublishDirOpts) error {
	panic("not implemented") // TODO: Implement
}

func (t *Tarball) Dependencies(ctx context.Context) ([]*pipeline.Artifact, error) {
	return []*pipeline.Artifact{
		t.Backend,
		t.Frontend,
		t.NPMPackages,
		t.BundledPlugins,
		t.Storybook,
	}, nil
}

func (t *Tarball) Filename(ctx context.Context) (string, error) {
	return packages.FileName(t.Name, t.Version, t.BuildID, t.Distribution, "tar.gz")
}
