package artifacts

import (
	"context"
	"log/slog"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/arguments"
	"github.com/grafana/grafana-build/backend"
	"github.com/grafana/grafana-build/docker"
	"github.com/grafana/grafana-build/flags"
	"github.com/grafana/grafana-build/packages"
	"github.com/grafana/grafana-build/pipeline"
)

var (
	DockerArguments = arguments.Join(
		TargzArguments,
		[]pipeline.Argument{
			arguments.DockerRegistry,
			arguments.DockerOrg,
			arguments.AlpineImage,
			arguments.UbuntuImage,
			arguments.TagFormat,
			arguments.UbuntuTagFormat,
			arguments.BoringTagFormat,
		},
	)
	DockerFlags = flags.JoinFlags(
		TargzFlags,
		flags.DockerFlags,
	)
)

var DockerInitializer = Initializer{
	InitializerFunc: NewDockerFromString,
	Arguments:       DockerArguments,
}

// PacakgeDocker uses a built tar.gz package to create a .rpm installer for RHEL-ish Linux distributions.
type Docker struct {
	Name       packages.Name
	Version    string
	BuildID    string
	Distro     backend.Distribution
	Enterprise bool

	Ubuntu       bool
	Registry     string
	Repositories []string
	Org          string
	BaseImage    string
	TagFormat    string

	Tarball *pipeline.Artifact
}

func (d *Docker) Dependencies(ctx context.Context) ([]*pipeline.Artifact, error) {
	return []*pipeline.Artifact{
		d.Tarball,
	}, nil
}

func (d *Docker) Builder(ctx context.Context, opts *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	targz, err := opts.Store.File(ctx, d.Tarball)
	if err != nil {
		return nil, err
	}

	return docker.Builder(opts.Client, opts.Client.Host().UnixSocket("/var/run/docker.sock"), targz), nil
}

func (d *Docker) BuildFile(ctx context.Context, builder *dagger.Container, opts *pipeline.ArtifactContainerOpts) (*dagger.File, error) {
	tags, err := docker.Tags(d.Org, d.Registry, d.Repositories, d.TagFormat, packages.NameOpts{
		Name:    d.Name,
		Version: d.Version,
		BuildID: d.BuildID,
		Distro:  d.Distro,
	})
	if err != nil {
		return nil, err
	}
	buildOpts := &docker.BuildOpts{
		// Tags are provided as the '-t' argument, and can include the registry domain as well as the repository.
		// Docker build supports building the same image with multiple tags.
		// You might want to also include a 'latest' version of the tag.
		Tags:      tags,
		Platform:  backend.Platform(d.Distro),
		BaseImage: d.BaseImage,
	}

	b := docker.Build(opts.Client, builder, buildOpts)

	return docker.Save(b, buildOpts), nil
}

func (d *Docker) BuildDir(ctx context.Context, builder *dagger.Container, opts *pipeline.ArtifactContainerOpts) (*dagger.Directory, error) {
	panic("not implemented") // TODO: Implement
}

func (d *Docker) Publisher(ctx context.Context, opts *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	panic("not implemented") // TODO: Implement
}

func (d *Docker) PublishFile(ctx context.Context, opts *pipeline.ArtifactPublishFileOpts) error {
	panic("not implemented") // TODO: Implement
}

func (d *Docker) PublisDir(ctx context.Context, opts *pipeline.ArtifactPublishDirOpts) error {
	panic("not implemented") // TODO: Implement
}

// Filename should return a deterministic file or folder name that this build will produce.
// This filename is used as a map key for caching, so implementers need to ensure that arguments or flags that affect the output
// also affect the filename to ensure that there are no collisions.
// For example, the backend for `linux/amd64` and `linux/arm64` should not both produce a `bin` folder, they should produce a
// `bin/linux-amd64` folder and a `bin/linux-arm64` folder. Callers can mount this as `bin` or whatever if they want.
func (d *Docker) Filename(ctx context.Context) (string, error) {
	ext := "docker.tar.gz"
	if d.Ubuntu {
		ext = "ubuntu.docker.tar.gz"
	}

	return packages.FileName(d.Name, d.Version, d.BuildID, d.Distro, ext)
}

func NewDockerFromString(ctx context.Context, log *slog.Logger, artifact string, state pipeline.StateHandler) (*pipeline.Artifact, error) {
	tarball, err := NewTarballFromString(ctx, log, artifact, state)
	if err != nil {
		return nil, err
	}

	options, err := pipeline.ParseFlags(artifact, DockerFlags)
	if err != nil {
		return nil, err
	}

	p, err := GetPackageDetails(ctx, options, state)
	if err != nil {
		return nil, err
	}

	ubuntu, err := options.Bool(flags.Ubuntu)
	if err != nil {
		return nil, err
	}

	// Ubuntu Version to use as the base for the Grafana docker image (if this is a ubuntu artifact)
	// This shouldn't fail if it's not set by the user, instead it'll default to 22.04 or something.
	ubuntuImage, err := state.String(ctx, arguments.UbuntuImage)
	if err != nil {
		return nil, err
	}

	// Same for Alpine
	alpineImage, err := state.String(ctx, arguments.AlpineImage)
	if err != nil {
		return nil, err
	}

	registry, err := state.String(ctx, arguments.DockerRegistry)
	if err != nil {
		return nil, err
	}

	org, err := state.String(ctx, arguments.DockerOrg)
	if err != nil {
		return nil, err
	}

	repos, err := options.StringSlice(flags.DockerRepositories)
	if err != nil {
		return nil, err
	}

	format, err := state.String(ctx, arguments.TagFormat)
	if err != nil {
		return nil, err
	}
	ubuntuFormat, err := state.String(ctx, arguments.UbuntuTagFormat)
	if err != nil {
		return nil, err
	}
	boringFormat, err := state.String(ctx, arguments.BoringTagFormat)
	if err != nil {
		return nil, err
	}

	base := alpineImage
	if ubuntu {
		format = ubuntuFormat
		base = ubuntuImage
	}

	if p.Name == packages.PackageEnterpriseBoring {
		format = boringFormat
	}

	log.Info("initializing Docker artifact", "Org", org, "registry", registry, "repos", repos, "tag", format)

	return pipeline.ArtifactWithLogging(ctx, log, &pipeline.Artifact{
		ArtifactString: artifact,
		Handler: &Docker{
			Name:       p.Name,
			Version:    p.Version,
			BuildID:    p.BuildID,
			Distro:     p.Distribution,
			Enterprise: p.Enterprise,
			Tarball:    tarball,

			Ubuntu:       ubuntu,
			BaseImage:    base,
			Registry:     registry,
			Org:          org,
			Repositories: repos,
			TagFormat:    format,
		},
		Type:  pipeline.ArtifactTypeFile,
		Flags: DockerFlags,
	})
}
