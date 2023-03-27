package containers

import (
	"github.com/grafana/grafana-build/executil"
)

type DistroBuildOptsFunc func(executil.Distribution, *BuildInfo) *executil.GoBuildOpts

var DefaultTags = []string{
	"netgo",
	"osusergo",
}

var DefaultBuildOpts = func(distro executil.Distribution, buildinfo *BuildInfo) *executil.GoBuildOpts {
	os, arch := executil.OSAndArch(distro)

	return &executil.GoBuildOpts{
		ExperimentalFlags: []string{},
		OS:                os,
		Arch:              arch,
		CGOEnabled:        true,
		TrimPath:          true,
		LDFlags: map[string][]string{
			"-X": buildinfo.LDFlags(),
		},
		Tags: DefaultTags,
	}
}

func BuildOptsStaticARM(distro executil.Distribution, buildinfo *BuildInfo) *executil.GoBuildOpts {
	var (
		os, arch = executil.OSAndArch(distro)
		arm      = executil.ArchVersion(distro)
	)

	return &executil.GoBuildOpts{
		ExperimentalFlags: []string{},
		OS:                os,
		Arch:              arch,
		GoARM:             executil.GoARM(arm),
		CGOEnabled:        true,
		TrimPath:          true,
		LDFlags: map[string][]string{
			"-w":                  nil,
			"-s":                  nil,
			"-X":                  buildinfo.LDFlags(),
			"-linkmode=external":  nil,
			"-extldflags=-static": nil,
		},
		Tags: DefaultTags,
	}
}

func BuildOptsStatic(distro executil.Distribution, buildinfo *BuildInfo) *executil.GoBuildOpts {
	var (
		os, arch = executil.OSAndArch(distro)
	)

	return &executil.GoBuildOpts{
		ExperimentalFlags: []string{},
		OS:                os,
		Arch:              arch,
		CGOEnabled:        true,
		TrimPath:          true,
		LDFlags: map[string][]string{
			"-w":                  nil,
			"-s":                  nil,
			"-X":                  buildinfo.LDFlags(),
			"-linkmode=external":  nil,
			"-extldflags=-static": nil,
		},
		Tags: DefaultTags,
	}
}

func BuildOptsDynamic(distro executil.Distribution, buildinfo *BuildInfo) *executil.GoBuildOpts {
	var (
		os, arch = executil.OSAndArch(distro)
	)

	return &executil.GoBuildOpts{
		ExperimentalFlags: []string{},
		OS:                os,
		Arch:              arch,
		CGOEnabled:        true,
		TrimPath:          true,
		LDFlags: map[string][]string{
			"-X": buildinfo.LDFlags(),
		},
		Tags: DefaultTags,
	}
}

func BuildOptsDynamicWindows(distro executil.Distribution, buildinfo *BuildInfo) *executil.GoBuildOpts {
	var (
		os, arch = executil.OSAndArch(distro)
	)

	return &executil.GoBuildOpts{
		ExperimentalFlags: []string{},
		OS:                os,
		Arch:              arch,
		BuildMode:         executil.BuildModeExe,
		CGOEnabled:        true,
		TrimPath:          true,
		LDFlags: map[string][]string{
			"-X": buildinfo.LDFlags(),
		},
		Tags: DefaultTags,
	}
}

var DistributionGoOpts = map[executil.Distribution]DistroBuildOptsFunc{
	executil.DistLinuxARM:   BuildOptsStaticARM,
	executil.DistLinuxARMv6: BuildOptsStaticARM,
	executil.DistLinuxARMv7: BuildOptsStaticARM,
	executil.DistLinuxARM64: BuildOptsStatic,
	executil.DistLinuxAMD64: BuildOptsStatic,

	executil.DistDarwinAMD64: BuildOptsDynamic,
	executil.DistDarwinARM64: BuildOptsDynamic,

	executil.DistWindowsAMD64: BuildOptsDynamicWindows,
	executil.DistWindowsARM64: BuildOptsDynamicWindows,
}
