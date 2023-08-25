package containers

import (
	"fmt"

	"github.com/grafana/grafana-build/executil"
)

type DistroBuildOptsFunc func(executil.Distribution, *BuildInfo) *executil.GoBuildOpts

var DefaultTags = []string{
	"netgo",
	"osusergo",
}

func ZigCC(distro executil.Distribution) string {
	target, ok := ZigTargets[distro]
	if !ok {
		target = "x86_64-linux-musl" // best guess? should probably retun an error but i don't want to
	}

	return fmt.Sprintf("zig cc -target %s", target)
}

func ZigCXX(distro executil.Distribution) string {
	target, ok := ZigTargets[distro]
	if !ok {
		target = "x86_64-linux-musl" // best guess? should probably retun an error but i don't want to
	}

	return fmt.Sprintf("zig c++ -target %s", target)
}

var DefaultBuildOpts = func(distro executil.Distribution, buildinfo *BuildInfo) *executil.GoBuildOpts {
	os, arch := executil.OSAndArch(distro)

	return &executil.GoBuildOpts{
		CC:                ZigCC(distro),
		CXX:               ZigCXX(distro),
		ExperimentalFlags: []string{},
		OS:                os,
		Arch:              arch,
		CGOEnabled:        true,
		TrimPath:          true,
		LDFlags: map[string][]string{
			"-X": buildinfo.LDFlags(),
		},
	}
}

// BuildOptsDynamicARM builds Grafana statically for the armv6/v7 architectures (not aarch64/arm64)
func BuildOptsDynamicARM(distro executil.Distribution, buildinfo *BuildInfo) *executil.GoBuildOpts {
	var (
		os, _ = executil.OSAndArch(distro)
		arm   = executil.ArchVersion(distro)
	)

	return &executil.GoBuildOpts{
		// CC:                ZigCC(distro),
		// CXX:               ZigCXX(distro),
		ExperimentalFlags: []string{},
		OS:                os,
		Arch:              "arm",
		GoARM:             executil.GoARM(arm),
		CGOEnabled:        true,
		TrimPath:          true,
		LDFlags: map[string][]string{
			"-X": buildinfo.LDFlags(),
		},
	}
}

func BuildOptsStatic(distro executil.Distribution, buildinfo *BuildInfo) *executil.GoBuildOpts {
	var (
		os, arch = executil.OSAndArch(distro)
	)

	return &executil.GoBuildOpts{
		CC:                ZigCC(distro),
		CXX:               ZigCXX(distro),
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
		CC:                ZigCC(distro),
		CXX:               ZigCXX(distro),
		ExperimentalFlags: []string{},
		OS:                os,
		Arch:              arch,
		CGOEnabled:        true,
		TrimPath:          true,
		LDFlags: map[string][]string{
			"-X": buildinfo.LDFlags(),
		},
	}
}

func BuildOptsDynamicDarwin(distro executil.Distribution, buildinfo *BuildInfo) *executil.GoBuildOpts {
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
	}
}

var ZigTargets = map[executil.Distribution]string{
	executil.DistLinuxAMD64:        "x86_64-linux-musl",
	executil.DistLinuxAMD64Dynamic: "x86_64-linux-musl",
	executil.DistLinuxARM64:        "aarch64-linux-musl",
	executil.DistLinuxARM64Dynamic: "aarch64-linux-musl",
	executil.DistLinuxARM:          "arm-linux-musleabihf",
	executil.DistLinuxARMv6:        "arm-linux-musleabihf",
	executil.DistLinuxARMv7:        "arm-linux-musleabihf",
}

var DistributionGoOpts = map[executil.Distribution]DistroBuildOptsFunc{
	// The Linux distros should all have an equivalent zig target in the ZigTargets map
	executil.DistLinuxARM:          BuildOptsDynamicARM,
	executil.DistLinuxARMv6:        BuildOptsDynamicARM,
	executil.DistLinuxARMv7:        BuildOptsDynamicARM,
	executil.DistLinuxARM64:        BuildOptsStatic,
	executil.DistLinuxARM64Dynamic: BuildOptsDynamic,
	executil.DistLinuxAMD64:        BuildOptsStatic,
	executil.DistLinuxAMD64Dynamic: BuildOptsDynamic,

	// Non-Linux distros can have whatever they want in CC and CXX; it'll get overridden
	// but it's probably not best to rely on that.
	executil.DistDarwinAMD64: BuildOptsDynamicDarwin,
	executil.DistDarwinARM64: BuildOptsDynamicDarwin,

	executil.DistWindowsAMD64: BuildOptsDynamicWindows,
	executil.DistWindowsARM64: BuildOptsDynamicWindows,

	executil.DistPlan9AMD64: BuildOptsDynamic,
}
