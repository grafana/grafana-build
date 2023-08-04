package pipelines

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
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

	// Run them in parallel
	for k, dir := range dirs {
		// Join the produced destination with the protocol given by the '--destination' flag.
		dst := strings.Join([]string{args.PublishOpts.Destination, k}, "/")
		grp.Go(PublishDirFunc(ctx, sm, d, dir, args.GCPOpts, dst))
	}

	return grp.Wait()
}

func ValidatePackageUpgrade(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return err
	}

	if len(packages) < 2 {
		return fmt.Errorf("at least two packages required for upgrade")
	}

	return validateUpgrade(ctx, d, packages, args.PackageInputOpts.Packages)
}

func ValidatePackageSignature(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return err
	}

	for i, name := range args.PackageInputOpts.Packages {
		pkg := packages[i]
		err := validateSignature(ctx, d, pkg, name, args.GPGOpts)
		if err != nil {
			return err
		}
	}

	return nil
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

	log.Printf("Validating docker image for v%s%s using platform %s\n", taropts.Version, taropts.Suffix, taropts.Distro)

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

	log.Printf("Validating deb package for v%s%s using debian:latest and platform %s\n", taropts.Version, taropts.Suffix, taropts.Distro)

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

	log.Printf("Validating rpm package for v%s%s using redhat/ubi8:latest and platform %s\n", taropts.Version, taropts.Suffix, taropts.Distro)

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

	log.Printf("Validating standalone tarball for v%s%s using ubuntu:22.10 and platform %s\n", taropts.Version, taropts.Suffix, taropts.Distro)

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

// validateLicense uses the given container and license path to validate the license for each edition (enterprise or oss)
func validateLicense(ctx context.Context, service *dagger.Container, licensePath string, taropts TarFileOpts) error {
	license, err := service.File(licensePath).Contents(ctx)
	if err != nil {
		return err
	}

	if taropts.Edition == "enterprise" {
		if !strings.Contains(license, "Grafana Enterprise") {
			return fmt.Errorf("license in package is not the Grafana Enterprise license agreement")
		}
	}

	if taropts.Edition == "" {
		if !strings.Contains(license, "GNU AFFERO GENERAL PUBLIC LICENSE") {
			return fmt.Errorf("license in package is not the Grafana open-source license agreement")
		}
	}

	return nil
}

// validateVersion uses the given container and version path to validate the version for each edition (enterprise or oss)
func validateVersion(ctx context.Context, service *dagger.Container, versionPath string, taropts TarFileOpts) error {
	version, err := service.File(versionPath).Contents(ctx)
	if err != nil {
		return err
	}

	if strings.TrimSpace(version) != taropts.Version {
		return fmt.Errorf("version in package does not match version in package name")
	}

	return nil
}

// validateUpgrade verifies the extension of the first package and proceeds with upgrade validation for the same extension
func validateUpgrade(ctx context.Context, d *dagger.Client, packages []*dagger.File, names []string) error {
	firstName := names[0]
	if filepath.Ext(firstName) == ".deb" {
		return validateDebUpgrade(ctx, d, packages, names)
	}

	if strings.HasSuffix(firstName, ".rpm") {
		return validateRpmUpgrade(ctx, d, packages, names)
	}

	return fmt.Errorf("invalid upgrade package extension")
}

// validateDebUpgrade receives a list of packages and package names, the names are used to retrieve information such as distro and edition
// the function expects all the packages to have the same distro, otherwise it outputs a distro mismatch error
// each package is installed to the same container and the license and version files are validated to see if the installation succeeded
func validateDebUpgrade(ctx context.Context, d *dagger.Client, packages []*dagger.File, names []string) error {
	var lastopts *TarFileOpts
	var container *dagger.Container
	for i, name := range names {
		if ext := filepath.Ext(name); ext != ".deb" {
			return fmt.Errorf("expected a file ending in .deb, received '%s'", ext)
		}

		pkg := packages[i]
		taropts := TarOptsFromFileName(name)
		if container == nil {
			container = d.Container(dagger.ContainerOpts{
				Platform: distroPlatform(taropts.Distro),
			}).From("debian:latest").
				WithExec([]string{"apt-get", "update"}).
				WithWorkdir("/usr/share/grafana")
		}

		if lastopts != nil {
			if lastopts.Distro != taropts.Distro {
				return fmt.Errorf("upgrade package distro mismatch")
			}

			log.Printf("Validating deb package upgrade from v%s%s to v%s%s using debian:latest and platform %s\n", lastopts.Version, lastopts.Suffix, taropts.Version, taropts.Suffix, lastopts.Distro)
		}

		container = container.
			WithFile("/src/package.deb", pkg).
			WithExec([]string{"apt-get", "install", "-y", "/src/package.deb"})

		if err := validateVersion(ctx, container, "/usr/share/grafana/VERSION", taropts); err != nil {
			return err
		}

		if err := validateLicense(ctx, container, "/usr/share/grafana/LICENSE", taropts); err != nil {
			return err
		}

		lastopts = &taropts
	}

	return nil
}

// validateRpmUpgrade receives a list of packages and package names, the names are used to retrieve information such as distro and edition
// the function expects all the packages to have the same distro, otherwise it outputs a distro mismatch error
// each package is installed to the same container and the license and version files are validated to see if the installation succeeded
func validateRpmUpgrade(ctx context.Context, d *dagger.Client, packages []*dagger.File, names []string) error {
	var lastopts *TarFileOpts
	var container *dagger.Container
	for i, name := range names {
		if ext := filepath.Ext(name); ext != ".rpm" {
			return fmt.Errorf("expected a file ending in .rpm, received '%s'", ext)
		}

		pkg := packages[i]
		taropts := TarOptsFromFileName(name)
		if container == nil {
			container = d.Container(dagger.ContainerOpts{
				Platform: distroPlatform(taropts.Distro),
			}).From("redhat/ubi8:latest").
				WithWorkdir("/usr/share/grafana")
		}

		if lastopts != nil {
			if lastopts.Distro != taropts.Distro {
				return fmt.Errorf("upgrade package distro mismatch")
			}

			log.Printf("Validating rpm package upgrade from v%s%s to v%s%s using redhat/ubi8:latest and platform %s\n", lastopts.Version, lastopts.Suffix, taropts.Version, taropts.Suffix, lastopts.Distro)
		}

		container = container.
			WithFile("/src/package.rpm", pkg).
			WithExec([]string{"yum", "install", "-y", "--allowerasing", "/src/package.rpm"})

		if err := validateVersion(ctx, container, "/usr/share/grafana/VERSION", taropts); err != nil {
			return err
		}

		if err := validateLicense(ctx, container, "/usr/share/grafana/LICENSE", taropts); err != nil {
			return err
		}

		lastopts = &taropts
	}

	return nil
}

// validateSignature uses the given package (rpm) and provided gpg public key to validate signature.
func validateSignature(ctx context.Context, d *dagger.Client, rpm *dagger.File, packageName string, opts *containers.GPGOpts) error {
	if ext := filepath.Ext(packageName); ext != ".rpm" {
		return fmt.Errorf("expected a .rpm file, received '%s'", ext)
	}

	var (
		taropts = TarOptsFromFileName(packageName)
	)

	log.Printf("Validating rpm package signature for v%s%s and platform %s\n", taropts.Version, taropts.Suffix, taropts.Distro)

	code, err := containers.RPMContainer(d, &containers.GPGOpts{Sign: true, GPGPublicKeyBase64: opts.GPGPublicKeyBase64}).
		WithFile("/src/package.rpm", rpm).
		WithExec([]string{"/bin/sh", "-c", "rpm --checksig /src/package.rpm | grep -qE 'digests signatures OK|pgp.+OK'"}).
		ExitCode(ctx)

	if err != nil || code != 0 {
		return fmt.Errorf("failed to validate gpg signature for rpm package")
	}
	return nil
}
