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

// BuildOptsStaticARM builds Grafana statically for the armv6/v7 architectures (not armhf/arm64)
func BuildOptsStaticARM(distro executil.Distribution, buildinfo *BuildInfo) *executil.GoBuildOpts {
	var (
		os, _ = executil.OSAndArch(distro)
		arm   = executil.ArchVersion(distro)
	)

	return &executil.GoBuildOpts{
		ExperimentalFlags: []string{},
		OS:                os,
		Arch:              "arm",
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

// BuildOptsDynamicARM builds Grafana statically for the armv6/v7 architectures (not armhf/arm64)
func BuildOptsDynamicARM(distro executil.Distribution, buildinfo *BuildInfo) *executil.GoBuildOpts {
	var (
		os, _ = executil.OSAndArch(distro)
		arm   = executil.ArchVersion(distro)
	)

	return &executil.GoBuildOpts{
		ExperimentalFlags: []string{},
		OS:                os,
		Arch:              "arm",
		GoARM:             executil.GoARM(arm),
		CGOEnabled:        true,
		TrimPath:          true,
		LDFlags: map[string][]string{
			"-X": buildinfo.LDFlags(),
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
	executil.DistLinuxARM:   BuildOptsDynamicARM,
	executil.DistLinuxARMv6: BuildOptsDynamicARM,
	executil.DistLinuxARMv7: BuildOptsDynamicARM,
	executil.DistLinuxARM64: BuildOptsStatic,
	executil.DistLinuxAMD64: BuildOptsStatic,

	executil.DistDarwinAMD64: BuildOptsDynamic,
	executil.DistDarwinARM64: BuildOptsDynamic,

	executil.DistWindowsAMD64: BuildOptsDynamicWindows,
	executil.DistWindowsARM64: BuildOptsDynamicWindows,

	executil.DistPlan9AMD64: BuildOptsDynamic,
}
