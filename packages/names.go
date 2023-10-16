package packages

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/grafana/grafana-build/executil"
)

type NameOpts struct {
	// Name is the name of the product in the package. 99% of the time, this will be "grafana" or "grafana-enterprise".
	Name      string
	Version   string
	BuildID   string
	Distro    executil.Distribution
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

var (
	ErrorNoName      = errors.New("all packages must have a name. Ex: \"grafana\", \"grafana-enteprrise\", \"grafana-rpi\"")
	ErrorNoVersion   = errors.New("all packages must have a version")
	ErrorNoBuildID   = errors.New("all packages must have a build ID")
	ErrorNoDistro    = errors.New("all packages must have a build distribution. Ex: \"linux/amd64\", \"darwin/arm64\", \"linux/plan9\"")
	ErrorNoExtension = errors.New("all packages must have a file extension. Ex: \"tar.gz\", \"deb\", \"docker.tar.gz\"")
)

// FileName returns a file name that matches this format: {grafana|grafana-enterprise}_{version}_{os}_{arch}_{build_number}.tar.gz
func FileName(opts NameOpts) (string, error) {
	if opts.Name == "" {
		return "", ErrorNoName
	}
	if opts.Version == "" {
		return "", ErrorNoVersion
	}
	if opts.BuildID == "" {
		return "", ErrorNoBuildID
	}
	if opts.Distro == "" {
		return "", ErrorNoDistro
	}
	var (
		name = opts.Name
		// This should return something like "linux", "arm"
		os, arch = executil.OSAndArch(opts.Distro)
		// If applicable this will be set to something like "7" (for arm7)
		archv = executil.ArchVersion(opts.Distro)
	)

	if archv != "" {
		arch = strings.Join([]string{arch, archv}, "-")
	}

	p := []string{name, opts.Version, opts.BuildID, os, arch}

	return fmt.Sprintf("%s.%s", strings.Join(p, "_"), opts.Extension), nil
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
		Name:    name,
		Version: version,
		BuildID: buildID,
		Distro:  executil.Distribution(strings.Join([]string{os, arch}, "/")),
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
