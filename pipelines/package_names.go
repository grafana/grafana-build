package pipelines

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/grafana/grafana-build/executil"
)

type TarFileOpts struct {
	Version string
	BuildID string
	// Edition is the flavor text after "grafana-", like "enterprise".
	Edition string
	Distro  executil.Distribution
}

func WithoutExt(name string) string {
	ext := filepath.Ext(name)
	n := strings.TrimSuffix(name, ext)

	// Explicitely handle `.gz` which might will also probably have a `.tar` extension as well.
	if ext == ".gz" {
		n = strings.TrimSuffix(n, ".ubuntu.docker.tar")
		n = strings.TrimSuffix(n, ".docker.tar")
		n = strings.TrimSuffix(n, ".tar")
	}

	return n
}

// TarFilename returns a file name that matches this format: {grafana|grafana-enterprise}_{version}_{os}_{arch}_{build_number}.tar.gz
func TarFilename(opts TarFileOpts) string {
	name := "grafana"
	if opts.Edition != "" {
		name = fmt.Sprintf("grafana-%s", opts.Edition)
	}

	var (
		// This should return something like "linux", "arm"
		os, arch = executil.OSAndArch(opts.Distro)
		// If applicable this will be set to something like "7" (for arm7)
		archv = executil.ArchVersion(opts.Distro)
	)

	if archv != "" {
		arch = strings.Join([]string{arch, archv}, "-")
	}

	p := []string{name, opts.Version, opts.BuildID, os, arch}

	return fmt.Sprintf("%s.tar.gz", strings.Join(p, "_"))
}

func TarOptsFromFileName(filename string) TarFileOpts {
	n := WithoutExt(filename)
	components := strings.Split(n, "_")
	if len(components) != 5 {
		return TarFileOpts{}
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
	edition := ""
	if n := strings.Split(name, "-"); len(n) != 1 {
		edition = n[1]
	}

	return TarFileOpts{
		Edition: edition,
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
