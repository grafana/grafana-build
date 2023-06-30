package pipelines

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// ValidatePackage downloads a package and validates from a Google Cloud Storage bucket.
func ValidatePackage(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return err
	}
	yarnCache := d.CacheVolume("yarn-cache")

	// Define all of the containers first, and where their artifacts will be exported to
	dirs := map[string]*dagger.Directory{}
	for i, name := range args.PackageInputOpts.Packages {
		pkg := packages[i]
		dir, err := validatePackage(ctx, d, pkg, src, yarnCache, name)
		if err != nil {
			return err
		}

		// replace .tar.gz with .e2e-artifacts/
		destination := DestinationName(name, "e2e-artifacts")
		dirs[destination] = dir
	}

	var (
		grp = &errgroup.Group{}
		sm  = semaphore.NewWeighted(args.ConcurrencyOpts.Parallel)
	)

	// Run them in parallel
	for k, dir := range dirs {
		// Join the produced destination with the protocol given by the '--destination' flag.
		dst := strings.Join([]string{args.PublishOpts.Destination, k}, "/")
		grp.Go(PublishDirFunc(ctx, sm, d, dir, args.GCPOpts, dst))
	}
	return grp.Wait()
}

func validatePackage(ctx context.Context, d *dagger.Client, pkg *dagger.File, src *dagger.Directory, yarnCache *dagger.CacheVolume, name string) (*dagger.Directory, error) {
	if strings.HasSuffix(name, ".docker.tar.gz") {
		return validateDocker(ctx, d, pkg, src, yarnCache, name)
	}

	if strings.HasSuffix(name, ".tar.gz") {
		return validateTarball(ctx, d, pkg, src, yarnCache, name)
	}

	if strings.HasSuffix(name, ".deb") {
		return validateDeb(ctx, d, pkg, src, yarnCache, name)
	}

	if strings.HasSuffix(name, ".rpm") {
		return validateRpm(ctx, d, pkg, src, yarnCache, name)
	}

	return nil, fmt.Errorf("unknown package extension")
}

// validateDocker uses the given package (.docker.tar.gz) and grafana source code (src) to run the e2e smoke tests.
// the returned directory is the e2e artifacts created by cypress (screenshots and videos).
func validateDocker(ctx context.Context, d *dagger.Client, pkg *dagger.File, src *dagger.Directory, yarnCache *dagger.CacheVolume, packageName string) (*dagger.Directory, error) {
	nodeVersion, err := containers.NodeVersion(d, src).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node version from source code: %w", err)
	}

	taropts := TarOptsFromFileName(packageName)

	// This grafana service runs in the background for the e2e tests
	service := d.Container(dagger.ContainerOpts{
		Platform: executil.Platform(taropts.Distro),
	}).Import(pkg).WithExposedPort(3000)

	// TODO: Add LICENSE to containers and implement validation

	return containers.ValidatePackage(d, service, src, yarnCache, nodeVersion), nil
}

// validateDeb uses the given package (deb) and grafana source code (src) to run the e2e smoke tests.
// the returned directory is the e2e artifacts created by cypress (screenshots and videos).
func validateDeb(ctx context.Context, d *dagger.Client, deb *dagger.File, src *dagger.Directory, yarnCache *dagger.CacheVolume, packageName string) (*dagger.Directory, error) {
	nodeVersion, err := containers.NodeVersion(d, src).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node version from source code: %w", err)
	}

	taropts := TarOptsFromFileName(packageName)

	// This grafana service runs in the background for the e2e tests
	serviceContainer := d.Container(dagger.ContainerOpts{
		Platform: executil.Platform(taropts.Distro),
	}).From("debian:latest").
		WithFile("/src/package.deb", deb).
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "/src/package.deb"}).
		WithWorkdir("/usr/share/grafana")

	if strings.Contains(packageName, "enterprise") {
		exitCode, err := serviceContainer.
			WithExec([]string{"grep", "-q", "Grafana Enterprise", "LICENSE"}).
			ExitCode(ctx)

		if err != nil || exitCode != 0 {
			return nil, fmt.Errorf("failed to validate enterprise license")
		}
	}

	// This grafana service runs in the background for the e2e tests
	service := serviceContainer.
		WithExec([]string{"grafana-server"}).
		WithExposedPort(3000)

	return containers.ValidatePackage(d, service, src, yarnCache, nodeVersion), nil
}

// validateRpm uses the given package (rpm) and grafana source code (src) to run the e2e smoke tests.
// the returned directory is the e2e artifacts created by cypress (screenshots and videos).
func validateRpm(ctx context.Context, d *dagger.Client, rpm *dagger.File, src *dagger.Directory, yarnCache *dagger.CacheVolume, packageName string) (*dagger.Directory, error) {
	nodeVersion, err := containers.NodeVersion(d, src).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node version from source code: %w", err)
	}

	taropts := TarOptsFromFileName(packageName)

	// This grafana service runs in the background for the e2e tests
	serviceContainer := d.Container(dagger.ContainerOpts{
		Platform: executil.Platform(taropts.Distro),
	}).From("redhat/ubi8:latest").
		WithFile("/src/package.rpm", rpm).
		WithExec([]string{"yum", "install", "-y", "/src/package.rpm"}).
		WithWorkdir("/usr/share/grafana")

	if strings.Contains(packageName, "enterprise") {
		exitCode, err := serviceContainer.
			WithExec([]string{"grep", "-q", "Grafana Enterprise", "LICENSE"}).
			ExitCode(ctx)

		if err != nil || exitCode != 0 {
			return nil, fmt.Errorf("failed to validate enterprise license")
		}
	}

	// This grafana service runs in the background for the e2e tests
	service := serviceContainer.
		WithExec([]string{"grafana-server"}).
		WithExposedPort(3000)

	return containers.ValidatePackage(d, service, src, yarnCache, nodeVersion), nil
}

// validateTarball uses the given package (pkg) and grafana source code (src) to run the e2e smoke tests.
// the returned directory is the e2e artifacts created by cypress (screenshots and videos).
func validateTarball(ctx context.Context, d *dagger.Client, pkg *dagger.File, src *dagger.Directory, yarnCache *dagger.CacheVolume, packageName string) (*dagger.Directory, error) {
	nodeVersion, err := containers.NodeVersion(d, src).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node version from source code: %w", err)
	}

	taropts := TarOptsFromFileName(packageName)
	archive := containers.ExtractedArchive(d, pkg, packageName)

	serviceContainer := d.Container(dagger.ContainerOpts{
		Platform: executil.Platform(taropts.Distro),
	}).From("ubuntu:22.10").
		WithExec([]string{"apt-get", "update", "-yq"}).
		WithExec([]string{"apt-get", "install", "-yq", "ca-certificates"}).
		WithDirectory("/src", archive).
		WithWorkdir("/src")

	if strings.Contains(packageName, "enterprise") {
		exitCode, err := serviceContainer.
			WithExec([]string{"grep", "-q", "Grafana Enterprise", "/src/LICENSE"}).
			ExitCode(ctx)

		if err != nil || exitCode != 0 {
			return nil, fmt.Errorf("failed to validate enterprise license")
		}
	}

	// This grafana service runs in the background for the e2e tests
	service := serviceContainer.
		WithExec([]string{"./bin/grafana", "server"}).
		WithExposedPort(3000)

	return containers.ValidatePackage(d, service, src, yarnCache, nodeVersion), nil
}
