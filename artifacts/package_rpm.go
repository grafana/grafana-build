package artifacts

import (
	"context"
	"log/slog"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/backend"
	"github.com/grafana/grafana-build/fpm"
	"github.com/grafana/grafana-build/packages"
	"github.com/grafana/grafana-build/pipeline"
)

var (
	RPMArguments = TargzArguments
	RPMFlags     = TargzFlags
)

var RPMInitializer = Initializer{
	InitializerFunc: NewRPMFromString,
	Arguments:       TargzArguments,
}

// PacakgeRPM uses a built tar.gz package to create a .rpm installer for RHEL-ish Linux distributions.
type RPM struct {
	Name         packages.Name
	Version      string
	BuildID      string
	Distribution backend.Distribution
	Enterprise   bool

	Tarball *pipeline.Artifact
}

func (d *RPM) Dependencies(ctx context.Context) ([]*pipeline.Artifact, error) {
	return []*pipeline.Artifact{
		d.Tarball,
	}, nil
}

func (d *RPM) Builder(ctx context.Context, opts *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	return fpm.Builder(opts.Client), nil
}

func (d *RPM) BuildFile(ctx context.Context, builder *dagger.Container, opts *pipeline.ArtifactContainerOpts) (*dagger.File, error) {
	targz, err := opts.Store.File(ctx, d.Tarball)
	if err != nil {
		return nil, err
	}

	return fpm.Build(builder, fpm.BuildOpts{
		Name:         d.Name,
		Enterprise:   d.Enterprise,
		Version:      d.Version,
		BuildID:      d.BuildID,
		Distribution: d.Distribution,
		PackageType:  fpm.PackageTypeRPM,
		ConfigFiles: [][]string{
			{"/src/packaging/rpm/sysconfig/grafana-server", "/pkg/etc/sysconfig/grafana-server"},
			{"/src/packaging/rpm/init.d/grafana-server", "/pkg/etc/init.d/grafana-server"},
			{"/src/packaging/rpm/systemd/grafana-server.service", "/pkg/usr/lib/systemd/system/grafana-server.service"},
		},
		AfterInstall: "/src/packaging/rpm/control/postinst",
		Depends: []string{
			"/sbin/service",
			"fontconfig",
			"freetype",
		},
		ExtraArgs: []string{
			"--rpm-posttrans=/src/packaging/rpm/control/posttrans",
			"--rpm-digest=sha256",
		},
		EnvFolder: "/pkg/etc/sysconfig",
	}, targz), nil
}

func (d *RPM) BuildDir(ctx context.Context, builder *dagger.Container, opts *pipeline.ArtifactContainerOpts) (*dagger.Directory, error) {
	panic("not implemented") // TODO: Implement
}

func (d *RPM) Publisher(ctx context.Context, opts *pipeline.ArtifactContainerOpts) (*dagger.Container, error) {
	panic("not implemented") // TODO: Implement
}

func (d *RPM) PublishFile(ctx context.Context, opts *pipeline.ArtifactPublishFileOpts) error {
	panic("not implemented") // TODO: Implement
}

func (d *RPM) PublisDir(ctx context.Context, opts *pipeline.ArtifactPublishDirOpts) error {
	panic("not implemented") // TODO: Implement
}

// Filename should return a deterministic file or folder name that this build will produce.
// This filename is used as a map key for caching, so implementers need to ensure that arguments or flags that affect the output
// also affect the filename to ensure that there are no collisions.
// For example, the backend for `linux/amd64` and `linux/arm64` should not both produce a `bin` folder, they should produce a
// `bin/linux-amd64` folder and a `bin/linux-arm64` folder. Callers can mount this as `bin` or whatever if they want.
func (d *RPM) Filename(ctx context.Context) (string, error) {
	return packages.FileName(d.Name, d.Version, d.BuildID, d.Distribution, "rpm")
}

func NewRPMFromString(ctx context.Context, log *slog.Logger, artifact string, state pipeline.StateHandler) (*pipeline.Artifact, error) {
	tarball, err := NewTarballFromString(ctx, log, artifact, state)
	if err != nil {
		return nil, err
	}
	options, err := pipeline.ParseFlags(artifact, RPMFlags)
	if err != nil {
		return nil, err
	}
	p, err := GetPackageDetails(ctx, options, state)
	if err != nil {
		return nil, err
	}
	return pipeline.ArtifactWithLogging(ctx, log, &pipeline.Artifact{
		ArtifactString: artifact,
		Handler: &RPM{
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
