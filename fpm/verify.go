package fpm

import (
	"context"
	"fmt"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/backend"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/e2e"
	"github.com/grafana/grafana-build/frontend"
	"github.com/grafana/grafana-build/gpg"
)

func VerifyDeb(ctx context.Context, d *dagger.Client, file *dagger.File, src *dagger.Directory, yarn *dagger.CacheVolume, distro backend.Distribution, enterprise bool) error {
	nodeVersion, err := frontend.NodeVersion(d, src).Stdout(ctx)
	if err != nil {
		return fmt.Errorf("failed to get node version from source code: %w", err)
	}

	var (
		platform = backend.Platform(distro)
	)

	// This grafana service runs in the background for the e2e tests
	service := d.Container(dagger.ContainerOpts{
		Platform: platform,
	}).From("debian:latest").
		WithFile("/src/package.deb", file).
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "/src/package.deb"}).
		WithWorkdir("/usr/share/grafana")

	if err := e2e.ValidateLicense(ctx, service, "/usr/share/grafana/LICENSE", enterprise); err != nil {
		return err
	}

	service = service.
		WithExec([]string{"grafana-server"}).
		WithExposedPort(3000)

	if _, err := containers.ExitError(ctx, e2e.ValidatePackage(d, service, src, yarn, nodeVersion)); err != nil {
		return err
	}

	return nil
}

func VerifyRpm(ctx context.Context, d *dagger.Client, file *dagger.File, src *dagger.Directory, yarn *dagger.CacheVolume, distro backend.Distribution, enterprise, sign bool, pubkey, privkey, passphrase string) error {
	nodeVersion, err := frontend.NodeVersion(d, src).Stdout(ctx)
	if err != nil {
		return fmt.Errorf("failed to get node version from source code: %w", err)
	}

	var (
		platform = backend.Platform(distro)
	)

	// This grafana service runs in the background for the e2e tests
	service := d.Container(dagger.ContainerOpts{
		Platform: platform,
	}).From("redhat/ubi8:latest").
		WithFile("/src/package.rpm", file).
		WithExec([]string{"yum", "install", "-y", "/src/package.rpm"}).
		WithWorkdir("/usr/share/grafana")

	if err := e2e.ValidateLicense(ctx, service, "/usr/share/grafana/LICENSE", enterprise); err != nil {
		return err
	}

	service = service.
		WithExec([]string{"grafana-server"}).
		WithExposedPort(3000)

	if _, err := containers.ExitError(ctx, e2e.ValidatePackage(d, service, src, yarn, nodeVersion)); err != nil {
		return err
	}
	if !sign {
		return nil
	}

	return gpg.VerifySignature(ctx, d, file, pubkey, privkey, passphrase)
}
