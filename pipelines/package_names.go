package pipelines

import (
	"fmt"
	"strings"

	"github.com/grafana/grafana-build/executil"
)

type TarFileOpts struct {
	Version      string
	BuildID      string
	IsEnterprise bool
	Distro       executil.Distribution
}

// TarFilename returns a file name that matches this format: {grafana|grafana-enterprise}_{version}_{os}_{arch}_{build_number}.tar.gz
func TarFilename(opts TarFileOpts) string {
	name := "grafana"
	if opts.IsEnterprise {
		name = "grafana-enterprise"
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

	p := []string{name, opts.Version, os, arch, opts.BuildID}

	return fmt.Sprintf("%s.tar.gz", strings.Join(p, "_"))
}

func TarOptsFromFileName(filename string) TarFileOpts {
	components := strings.Split(strings.TrimSuffix(filename, ".tar.gz"), "_")
	if len(components) != 5 {
		return TarFileOpts{}
	}

	var (
		name    = components[0]
		version = components[1]
		os      = components[2]
		arch    = components[3]
		buildID = components[4]
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

	return TarFileOpts{
		IsEnterprise: name == "grafana-enterprise",
		Version:      version,
		BuildID:      buildID,
		Distro:       executil.Distribution(strings.Join([]string{os, arch}, "/")),
	}
}
