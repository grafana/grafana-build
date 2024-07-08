package artifacts

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

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
			arguments.ProDirectory,
			arguments.DockerRegistry,
			arguments.DockerOrg,
			arguments.AlpineImage,
			arguments.UbuntuImage,
			arguments.TagFormat,
			arguments.UbuntuTagFormat,
			arguments.BoringTagFormat,

			arguments.ProDockerRegistry,
			arguments.ProDockerOrg,
			arguments.ProDockerRepo,
			arguments.ProTagFormat,
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
	Pro        bool
	ProDir     *dagger.Directory

	Ubuntu       bool
	Registry     string
	Repositories []string
	Org          string
	BaseImage    string
	TagFormat    string

	// ProRegistry is the docker registry when using the `pro` name. (e.g. hub.docker.io)
	ProRegistry string
	// ProOrg is the docker org when using the `pro` name. (e.g. grafana)
	ProOrg string
	// ProOrg is the docker repo when using the `pro` name. (e.g. grafana-pro)
	ProRepo string
	// ProTagFormat is the docker tag format when using the `pro` name. (e.g. {{ .version }}-{{ .os }}-{{ .arch }})
	ProTagFormat string

	Tarball *pipeline.Artifact

	// Building the Pro image requires a Debian package instead of a tar.gz
	Deb *pipeline.Artifact

	// Src is the Grafana source code for running e2e tests when validating.
	// The grafana source should not be used for anything else when building a docker image. All files in the Docker image, including the Dockerfile, should be
	// from the tar.gz file.
	Src       *dagger.Directory
	YarnCache *dagger.CacheVolume
}

func (d *Docker) Dependencies(ctx context.Context) ([]*pipeline.Artifact, error) {
	if d.Pro {
		return []*pipeline.Artifact{
			d.Deb,
		}, nil
	}

	return []*pipeline.Artifact{
		d.Tarball,
	}, nil
}

func (d *Docker) proBuilder(ctx context.Context, opts *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	deb, err := opts.Store.File(ctx, d.Deb)
	if err != nil {
		return nil, fmt.Errorf("error getting deb from state: %w", err)
	}

	socket := opts.Client.Host().UnixSocket("/var/run/docker.sock")

	return opts.Client.Container().From("docker").
		WithUnixSocket("/var/run/docker.sock", socket).
		WithMountedDirectory("/src", d.ProDir).
		WithMountedFile("/src/grafana.deb", deb).
		WithWorkdir("/src"), nil
}

func (d *Docker) Builder(ctx context.Context, opts *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	if d.Pro {
		return d.proBuilder(ctx, opts)
	}

	targz, err := opts.Store.File(ctx, d.Tarball)
	if err != nil {
		return nil, err
	}

	return docker.Builder(opts.Client, opts.Client.Host().UnixSocket("/var/run/docker.sock"), targz), nil
}

func (d *Docker) buildPro(ctx context.Context, builder *dagger.Container, opts *pipeline.ArtifactContainerOpts) (*dagger.File, error) {
	tags, err := docker.Tags(d.ProOrg, d.ProRegistry, []string{d.ProRepo}, d.ProTagFormat, packages.NameOpts{
		Name:    d.Name,
		Version: d.Version,
		BuildID: d.BuildID,
		Distro:  d.Distro,
	})

	if err != nil {
		return nil, err
	}

	builder = docker.Build(opts.Client, builder, &docker.BuildOpts{
		Dockerfile: "./docker/hosted-grafana-all/Dockerfile",
		Tags:       tags,
		Platform:   dagger.Platform("linux/amd64"),
		BuildArgs: []string{
			"RELEASE_TYPE=prerelease",
			// I think because deb files use a ~ as a version delimiter of some kind, so the hg docker image uses that instead of a -
			fmt.Sprintf("GRAFANA_VERSION=%s", strings.Replace(d.Version, "-", "~", 1)),
		},
	})

	// Save the resulting docker image to the local filesystem
	return builder.WithExec([]string{"docker", "save", tags[0], "-o", "pro.tar"}).File("pro.tar"), nil
}

func (d *Docker) BuildFile(ctx context.Context, builder *dagger.Container, opts *pipeline.ArtifactContainerOpts) (*dagger.File, error) {
	if d.Pro {
		return d.buildPro(ctx, builder, opts)
	}

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
		Tags:     tags,
		Platform: backend.Platform(d.Distro),
		BuildArgs: []string{
			"GRAFANA_TGZ=grafana.tar.gz",
			"GO_SRC=tgz-builder",
			"JS_SRC=tgz-builder",
			fmt.Sprintf("BASE_IMAGE=%s", d.BaseImage),
		},
	}

	b := docker.Build(opts.Client, builder, buildOpts)

	return docker.Save(b, buildOpts), nil
}

func (d *Docker) BuildDir(ctx context.Context, builder *dagger.Container, opts *pipeline.ArtifactContainerOpts) (*dagger.Directory, error) {
	panic("This artifact does not produce directories")
}

func (d *Docker) publishPro(ctx context.Context, opts *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	panic("not implemented")
}

func (d *Docker) Publisher(ctx context.Context, opts *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	if d.Pro {
		return d.publishPro(ctx, opts)
	}
	socket := opts.Client.Host().UnixSocket("/var/run/docker.sock")
	return opts.Client.Container().From("docker").WithUnixSocket("/var/run/docker.sock", socket), nil
}

func (d *Docker) PublishFile(ctx context.Context, opts *pipeline.ArtifactPublishFileOpts) error {
	panic("not implemented")
}

func (d *Docker) PublisDir(ctx context.Context, opts *pipeline.ArtifactPublishDirOpts) error {
	panic("This artifact does not produce directories")
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

func (d *Docker) VerifyFile(ctx context.Context, client *dagger.Client, file *dagger.File) error {
	// Currently verifying riscv64 is unsupported (because alpine and ubuntu don't have riscv64 images yet)
	if _, arch := backend.OSAndArch(d.Distro); arch == "riscv64" {
		return nil
	}

	if d.Pro {
		return nil
	}

	return docker.Verify(ctx, client, file, d.Src, d.YarnCache, d.Distro)
}

func (d *Docker) VerifyDirectory(ctx context.Context, client *dagger.Client, dir *dagger.Directory) error {
	panic("not implemented") // TODO: Implement
}

func NewDockerFromString(ctx context.Context, log *slog.Logger, artifact string, state pipeline.StateHandler) (*pipeline.Artifact, error) {
	var (
		pro     bool
		tarball *pipeline.Artifact
		deb     *pipeline.Artifact
	)

	options, err := pipeline.ParseFlags(artifact, DockerFlags)
	if err != nil {
		return nil, err
	}

	p, err := GetPackageDetails(ctx, options, state)
	if err != nil {
		return nil, err
	}

	if p.Name == packages.PackagePro {
		pro = true
	}

	// The Pro docker image depends on the deb while everything else depends on the targz
	if pro {
		artifact, err := NewDebFromString(ctx, log, artifact, state)
		if err != nil {
			return nil, err
		}
		deb = artifact
	} else {
		artifact, err := NewTarballFromString(ctx, log, artifact, state)
		if err != nil {
			return nil, err
		}

		tarball = artifact
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
	proRegistry, err := state.String(ctx, arguments.ProDockerRegistry)
	if err != nil {
		return nil, err
	}
	proOrg, err := state.String(ctx, arguments.ProDockerOrg)
	if err != nil {
		return nil, err
	}
	proRepo, err := state.String(ctx, arguments.ProDockerRepo)
	if err != nil {
		return nil, err
	}
	proTagFormat, err := state.String(ctx, arguments.ProTagFormat)
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

	var (
		proDir *dagger.Directory
	)
	if pro {
		dir, err := state.Directory(ctx, arguments.ProDirectory)
		if err != nil {
			return nil, err
		}

		proDir = dir
	}

	src, err := state.Directory(ctx, arguments.GrafanaDirectory)
	if err != nil {
		return nil, err
	}

	yarnCache, err := state.CacheVolume(ctx, arguments.YarnCacheDirectory)
	if err != nil {
		return nil, err
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
			Pro:        pro,
			ProDir:     proDir,
			Tarball:    tarball,
			Deb:        deb,

			Ubuntu:       ubuntu,
			BaseImage:    base,
			Registry:     registry,
			Org:          org,
			Repositories: repos,
			TagFormat:    format,

			ProRegistry:  proRegistry,
			ProOrg:       proOrg,
			ProRepo:      proRepo,
			ProTagFormat: proTagFormat,

			Src:       src,
			YarnCache: yarnCache,
		},
		Type:  pipeline.ArtifactTypeFile,
		Flags: DockerFlags,
	})
}
