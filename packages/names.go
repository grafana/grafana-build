package packages

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/grafana/grafana-build/backend"
)

type Name string

const (
	PackageGrafana          Name = "grafana"
	PackageEnterprise       Name = "grafana-enterprise"
	PackageEnterpriseBoring Name = "grafana-enterprise-boringcrypto"
	PackagePro              Name = "grafana-pro"
	PackageNightly          Name = "grafana-nightly"
)

type NameOpts struct {
	// Name is the name of the product in the package. 99% of the time, this will be "grafana" or "grafana-enterprise".
	Name      Name
	Version   string
	BuildID   string
	Distro    backend.Distribution
	Extension string
}

// WithoutExt removes the file extension from the given package string.
// It is aware of multiple-period exnteions, like ".docker.tar.gz", ".ubuntu.deb", etc.
func WithoutExt(name string) string {
	cmp := strings.Split(name, "_")
	last := cmp[len(cmp)-1]
	exti := strings.Index(last, ".")
	lastWithoutExt := last[:exti]

	return strings.Join(append(cmp[:len(cmp)-1], lastWithoutExt), "_")
}

// FileName returns a file name that matches this format: {grafana|grafana-enterprise}_{version}_{os}_{arch}_{build_number}.tar.gz
func FileName(name Name, version, buildID string, distro backend.Distribution, extension string) (string, error) {
	var (
		// This should return something like "linux", "arm"
		os, arch = backend.OSAndArch(distro)
		// If applicable this will be set to something like "7" (for arm7)
		archv = backend.ArchVersion(distro)
	)

	if archv != "" {
		arch = strings.Join([]string{arch, archv}, "-")
	}

	p := []string{string(name), version, buildID, os, arch}

	return fmt.Sprintf("%s.%s", strings.Join(p, "_"), extension), nil
}

func NameOptsFromFileName(filename string) NameOpts {
	filename = filepath.Base(filename)
	n := WithoutExt(filename)
	components := strings.Split(n, "_")
	if len(components) != 5 {
		return NameOpts{}
	}

	var (
		name    = components[0]
		version = components[1]
		buildID = components[2]
		os      = components[3]
		arch    = components[4]
	)

	if archv := strings.Split(arch, "-"); len(archv) == 2 {
		// The reverse operation of removing the 'v' for 'arm' because the golang environment variable
		// GOARM doesn't want it, but the docker --platform flag either doesn't care or does want it.
		if archv[0] == "arm" {
			archv[1] = "v" + archv[1]
		}

		// arm-7 should become arm/v7
		arch = strings.Join([]string{archv[0], archv[1]}, "/")
	}

	return NameOpts{
		Name:    Name(name),
		Version: version,
		BuildID: buildID,
		Distro:  backend.Distribution(strings.Join([]string{os, arch}, "/")),
	}
}

// ReplaceExt replaces the extension of the given package name.
// For example, if the input package name (original) is grafana_v1.0.0_linux-arm64_1.tar.gz, then
// derivatives should have the same name, but with a different extension.
// This function also removes the leading directory and removes the URL scheme prefix.
func ReplaceExt(original, extension string) string {
	n := strings.TrimPrefix(WithoutExt(original), "file://")
	if extension == "" {
		return filepath.Base(n)
	}

	return filepath.Base(fmt.Sprintf("%s.%s", n, extension))
}
