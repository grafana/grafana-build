package pipelines

import (
	"context"
	"fmt"
	"log"
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
		destination := name + ".e2e-artifacts"
		dirs[destination] = dir
	}

	var (
		grp = &errgroup.Group{}
		sm  = semaphore.NewWeighted(args.ConcurrencyOpts.Parallel)
	)

	log.Println("Parallel:", args.ConcurrencyOpts.Parallel)
	log.Println("Parallel:", args.ConcurrencyOpts.Parallel)
	log.Println("Parallel:", args.ConcurrencyOpts.Parallel)
	log.Println("Parallel:", args.ConcurrencyOpts.Parallel)
	// Run them in parallel
	for k, dir := range dirs {
		// Join the produced destination with the protocol given by the '--destination' flag.
		dst := strings.Join([]string{args.PublishOpts.Destination, k}, "/")
		grp.Go(PublishDirFunc(ctx, sm, d, dir, args.GCPOpts, dst))
	}

	return grp.Wait()
}

func distroPlatform(distro executil.Distribution) dagger.Platform {
	platform := executil.Platform(distro)
	if _, arch := executil.OSAndArch(distro); arch == "arm" {
		// armv7 and armv6 just use armv7
		return dagger.Platform("linux/arm/v7")
	}

	return platform
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

	var (
		taropts  = TarOptsFromFileName(packageName)
		platform = distroPlatform(taropts.Distro)
	)

	log.Printf("Validating docker image for v%s-%s using platform %s\n", taropts.Version, taropts.Edition, taropts.Distro)

	// This grafana service runs in the background for the e2e tests
	// Just guessing that maybe we need to add the "PACKAGE" environment variable here to prevent weird caching collisions
	// BuildKit should be smart enough to know that it's a different docker image though
	service := d.Container(dagger.ContainerOpts{
		Platform: platform,
	}).
		Import(pkg).
		WithEnvVariable("PACKAGE", packageName).
		WithExposedPort(3000)

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

	var (
		taropts  = TarOptsFromFileName(packageName)
		platform = distroPlatform(taropts.Distro)
	)

	log.Printf("Validating deb package for v%s-%s using debian:latest and platform %s\n", taropts.Version, taropts.Edition, taropts.Distro)

	// This grafana service runs in the background for the e2e tests
	service := d.Container(dagger.ContainerOpts{
		Platform: platform,
	}).From("debian:latest").
		WithFile("/src/package.deb", deb).
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "/src/package.deb"}).
		WithWorkdir("/usr/share/grafana")

	err = validateLicense(ctx, service, "/usr/share/grafana/LICENSE", taropts)
	if err != nil {
		return nil, err
	}

	service = service.
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

	var (
		taropts  = TarOptsFromFileName(packageName)
		platform = distroPlatform(taropts.Distro)
	)

	log.Printf("Validating rpm package for v%s-%s using redhat/ubi8:latest and platform %s\n", taropts.Version, taropts.Edition, taropts.Distro)

	// This grafana service runs in the background for the e2e tests
	service := d.Container(dagger.ContainerOpts{
		Platform: platform,
	}).From("redhat/ubi8:latest").
		WithFile("/src/package.rpm", rpm).
		WithExec([]string{"yum", "install", "-y", "/src/package.rpm"}).
		WithWorkdir("/usr/share/grafana")

	err = validateLicense(ctx, service, "/usr/share/grafana/LICENSE", taropts)
	if err != nil {
		return nil, err
	}

	service = service.
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

	var (
		taropts  = TarOptsFromFileName(packageName)
		platform = distroPlatform(taropts.Distro)
		archive  = containers.ExtractedArchive(d, pkg, packageName)
	)

	log.Printf("Validating standalone tarball for v%s-%s using ubuntu:22.10 and platform %s\n", taropts.Version, taropts.Edition, taropts.Distro)

	// This grafana service runs in the background for the e2e tests
	service := d.Container(dagger.ContainerOpts{
		Platform: platform,
	}).From("ubuntu:22.10").
		WithExec([]string{"apt-get", "update", "-yq"}).
		WithExec([]string{"apt-get", "install", "-yq", "ca-certificates", "libfontconfig1"}).
		WithDirectory("/src", archive).
		WithWorkdir("/src")

	err = validateLicense(ctx, service, "/src/LICENSE", taropts)
	if err != nil {
		return nil, err
	}

	service = service.
		WithExec([]string{"./bin/grafana", "server"}).
		WithExposedPort(3000)

	return containers.ValidatePackage(d, service, src, yarnCache, nodeVersion), nil
}

// validateLicense uses the given service and license path to validate the license for each edition (enterprise or oss)
func validateLicense(ctx context.Context, service *dagger.Container, licensePath string, taropts TarFileOpts) error {
	license, err := service.File(licensePath).Contents(ctx)
	if taropts.Edition == "enterprise" {
		if err != nil || !strings.Contains(license, "Grafana Enterprise") {
			return fmt.Errorf("failed to validate enterprise license")
		}
	}

	if taropts.Edition == "" {
		if err != nil || !strings.Contains(license, "GNU AFFERO GENERAL PUBLIC LICENSE") {
			return fmt.Errorf("failed to validate open-source license")
		}
	}

	return nil
}
