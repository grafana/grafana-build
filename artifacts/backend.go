package artifacts

import (
	"context"
	"log/slog"
	"path"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/arguments"
	"github.com/grafana/grafana-build/backend"
	"github.com/grafana/grafana-build/flags"
	"github.com/grafana/grafana-build/packages"
	"github.com/grafana/grafana-build/pipeline"
)

const BackendKey = "backend"

var (
	BackendArguments = []pipeline.Argument{
		arguments.GrafanaDirectory,
		arguments.EnterpriseDirectory,
		arguments.GoVersion,
		arguments.ViceroyVersion,
	}

	BackendFlags = flags.JoinFlags(
		flags.PackageNameFlags,
		flags.DistroFlags(),
	)
)

type Backend struct {
	// Name allows different backend compilations to be different even if all other factors are the same.
	// For example, Grafana Enterprise, Grafana, and Grafana Pro may be built using the same options,
	// but are fundamentally different because of the source code of the bianry.
	Name           packages.Name
	Src            *dagger.Directory
	Distribution   backend.Distribution
	BuildOpts      *backend.BuildOpts
	GoVersion      string
	ViceroyVersion string

	// Version is embedded in the binary at build-time
	Version string
}

func (b *Backend) Builder(ctx context.Context, opts *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	return backend.Builder(
		opts.Client,
		opts.Log,
		b.Distribution,
		b.BuildOpts,
		opts.Platform,
		b.Src,
		b.GoVersion,
		b.ViceroyVersion,
	)
}

func (b *Backend) Dependencies(ctx context.Context) ([]*pipeline.Artifact, error) {
	return nil, nil
}

func (b *Backend) BuildFile(ctx context.Context, builder *dagger.Container, opts *pipeline.ArtifactContainerOpts) (*dagger.File, error) {
	panic("not implemented") // TODO: Implement
}

func (b *Backend) BuildDir(ctx context.Context, builder *dagger.Container, opts *pipeline.ArtifactContainerOpts) (*dagger.Directory, error) {
	f, err := b.Filename(ctx)
	if err != nil {
		return nil, err
	}

	return backend.Build(
		builder,
		b.Src,
		b.Distribution,
		f,
		b.BuildOpts,
	), nil
}

func (b *Backend) Publisher(ctx context.Context, opts *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	panic("not implemented") // TODO: Implement
}

func (b *Backend) PublishFile(ctx context.Context, opts *pipeline.ArtifactPublishFileOpts) error {
	panic("not implemented") // TODO: Implement
}

func (b *Backend) PublisDir(ctx context.Context, opts *pipeline.ArtifactPublishDirOpts) error {
	panic("not implemented") // TODO: Implement
}

// Filename should return a deterministic file or folder name that this build will produce.
// This filename is used as a map key for caching, so implementers need to ensure that arguments or flags that affect the output
// also affect the filename to ensure that there are no collisions.
// For example, the backend for `linux/amd64` and `linux/arm64` should not both produce a `bin` folder, they should produce a
// `bin/linux-amd64` folder and a `bin/linux-arm64` folder. Callers can mount this as `bin` or whatever if they want.
func (b *Backend) Filename(ctx context.Context) (string, error) {
	return path.Join("bin", string(b.Name), string(b.Distribution)), nil
}

type NewBackendOpts struct {
	Name           packages.Name
	Src            *dagger.Directory
	Distribution   backend.Distribution
	GoVersion      string
	ViceroyVersion string
	Version        string
	Experiments    []string
	Tags           []string
	Static         bool
	WireTag        string
}

func NewBackend(ctx context.Context, log *slog.Logger, artifact string, opts *NewBackendOpts) (*pipeline.Artifact, error) {
	bopts := &backend.BuildOpts{
		Version:           opts.Version,
		ExperimentalFlags: opts.Experiments,
		Tags:              opts.Tags,
		Static:            opts.Static,
		WireTag:           opts.WireTag,
	}

	log.Info("Initializing backend artifact with options", "static", opts.Static, "version", opts.Version, "name", opts.Name, "distro", opts.Distribution)
	return pipeline.ArtifactWithLogging(ctx, log, &pipeline.Artifact{
		ArtifactString: artifact,
		Type:           pipeline.ArtifactTypeDirectory,
		Flags:          BackendFlags,
		Handler: &Backend{
			Name:           opts.Name,
			Distribution:   opts.Distribution,
			BuildOpts:      bopts,
			GoVersion:      opts.GoVersion,
			ViceroyVersion: opts.ViceroyVersion,
			Src:            opts.Src,
		},
	})
}
