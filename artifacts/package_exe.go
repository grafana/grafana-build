package artifacts

import (
	"context"
	"fmt"
	"log/slog"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/backend"
	"github.com/grafana/grafana-build/exe"
	"github.com/grafana/grafana-build/packages"
	"github.com/grafana/grafana-build/pipeline"
)

var (
	ExeArguments = TargzArguments
	ExeFlags     = TargzFlags
)

var ExeInitializer = Initializer{
	InitializerFunc: NewExeFromString,
	Arguments:       TargzArguments,
}

// PacakgeExe uses a built tar.gz package to create a .exe installer for exeian based Linux distributions.
type Exe struct {
	Name         packages.Name
	Version      string
	BuildID      string
	Distribution backend.Distribution
	Enterprise   bool

	Tarball *pipeline.Artifact
}

func (d *Exe) Dependencies(ctx context.Context) ([]*pipeline.Artifact, error) {
	return []*pipeline.Artifact{
		d.Tarball,
	}, nil
}

func (d *Exe) Builder(ctx context.Context, opts *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	return exe.Builder(opts.Client)
}

func (d *Exe) BuildFile(ctx context.Context, builder *dagger.Container, opts *pipeline.ArtifactContainerOpts) (*dagger.File, error) {
	targz, err := opts.Store.File(ctx, d.Tarball)
	if err != nil {
		return nil, err
	}

	return exe.Build(opts.Client, builder, targz, d.Enterprise), nil
}

func (d *Exe) BuildDir(ctx context.Context, builder *dagger.Container, opts *pipeline.ArtifactContainerOpts) (*dagger.Directory, error) {
	// Not a directory so this shouldn't be called
	return nil, nil
}

func (d *Exe) Publisher(ctx context.Context, opts *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	return nil, nil
}

func (d *Exe) PublishFile(ctx context.Context, opts *pipeline.ArtifactPublishFileOpts) error {
	panic("not implemented") // TODO: Implement
}

func (d *Exe) PublishDir(ctx context.Context, opts *pipeline.ArtifactPublishDirOpts) error {
	// Not a directory so this shouldn't be called
	return nil
}

// Filename should return a deterministic file or folder name that this build will produce.
// This filename is used as a map key for caching, so implementers need to ensure that arguments or flags that affect the output
// also affect the filename to ensure that there are no collisions.
// For example, the backend for `linux/amd64` and `linux/arm64` should not both produce a `bin` folder, they should produce a
// `bin/linux-amd64` folder and a `bin/linux-arm64` folder. Callers can mount this as `bin` or whatever if they want.
func (d *Exe) Filename(ctx context.Context) (string, error) {
	return packages.FileName(d.Name, d.Version, d.BuildID, d.Distribution, "exe")
}

func (d *Exe) VerifyFile(ctx context.Context, client *dagger.Client, file *dagger.File) error {
	return nil
}

func (d *Exe) VerifyDirectory(ctx context.Context, client *dagger.Client, dir *dagger.Directory) error {
	panic("not implemented") // TODO: Implement
}

func NewExeFromString(ctx context.Context, log *slog.Logger, artifact string, state pipeline.StateHandler) (*pipeline.Artifact, error) {
	tarball, err := NewTarballFromString(ctx, log, artifact, state)
	if err != nil {
		return nil, err
	}
	options, err := pipeline.ParseFlags(artifact, ExeFlags)
	if err != nil {
		return nil, err
	}
	p, err := GetPackageDetails(ctx, options, state)
	if err != nil {
		return nil, err
	}

	if !backend.IsWindows(p.Distribution) {
		return nil, fmt.Errorf("distribution ('%s') for exe '%s' is not a Windows distribution", string(p.Distribution), artifact)
	}

	return pipeline.ArtifactWithLogging(ctx, log, &pipeline.Artifact{
		ArtifactString: artifact,
		Handler: &Exe{
			Name:         p.Name,
			Version:      p.Version,
			BuildID:      p.BuildID,
			Distribution: p.Distribution,
			Enterprise:   p.Enterprise,
			Tarball:      tarball,
		},
		Type:  pipeline.ArtifactTypeFile,
		Flags: TargzFlags,
	})
}
