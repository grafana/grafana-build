package executil

import (
	"fmt"
	"os/exec"
	"strings"
)

// Distribution is a string that represents the GOOS and GOARCH environment variables joined by a "/".
// Optionally, if there is an extra argument specific to that architecture, it will be the last segment of the string.
// Examples:
// - "linux/arm/v6" = GOOS=linux, GOARCH=arm, GOARM=6
// - "linux/arm/v7" = GOOS=linux, GOARCH=arm, GOARM=7
// - "linux/amd64/v7" = GOOS=linux, GOARCH=arm, GOARM=7
// - "linux/amd64/v2" = GOOS=linux, GOARCH=amd64, GOAMD64=v2
// The list of distributions is built from the command "go tool dist list".
// While not all are used, it at least represents the possible combinations.
type Distribution string

const (
	DistDarwinAMD64   Distribution = "darwin/amd64"
	DistDarwinAMD64v1 Distribution = "darwin/amd64/v1"
	DistDarwinAMD64v2 Distribution = "darwin/amd64/v2"
	DistDarwinAMD64v3 Distribution = "darwin/amd64/v3"
	DistDarwinAMD64v4 Distribution = "darwin/amd64/v4"
	DistDarwinARM64   Distribution = "darwin/arm64"
)

const (
	DistFreeBSD386          Distribution = "freebsd/386"
	DistFreeBSD386SSE2      Distribution = "freebsd/386/sse2"
	DistFreeBSD386SoftFloat Distribution = "freebsd/386/softfloat"
	DistFreeBSDAMD64        Distribution = "freebsd/amd64"
	DistFreeBSDAMD64v1      Distribution = "freebsd/amd64/v1"
	DistFreeBSDAMD64v2      Distribution = "freebsd/amd64/v2"
	DistFreeBSDAMD64v3      Distribution = "freebsd/amd64/v3"
	DistFreeBSDAMD64v4      Distribution = "freebsd/amd64/v4"
	DistFreeBSDARM          Distribution = "freebsd/arm"
	DistFreeBSDARM64        Distribution = "freebsd/arm64"
	DistFreeBSDRISCV        Distribution = "freebsd/riscv64"
)

const (
	DistIllumosAMD64   Distribution = "illumos/amd64"
	DistIllumosAMD64v1 Distribution = "illumos/amd64/v1"
	DistIllumosAMD64v2 Distribution = "illumos/amd64/v2"
	DistIllumosAMD64v3 Distribution = "illumos/amd64/v3"
	DistIllumosAMD64v4 Distribution = "illumos/amd64/v4"
)
const (
	DistLinux386          Distribution = "linux/386"
	DistLinux386SSE2      Distribution = "linux/386/sse2"
	DistLinux386SoftFloat Distribution = "linux/386/softfloat"
	DistLinuxAMD64        Distribution = "linux/amd64"
	DistLinuxAMD64v1      Distribution = "linux/amd64/v1"
	DistLinuxAMD64v2      Distribution = "linux/amd64/v2"
	DistLinuxAMD64v3      Distribution = "linux/amd64/v3"
	DistLinuxAMD64v4      Distribution = "linux/amd64/v4"
	DistLinuxARM          Distribution = "linux/arm"
	DistLinuxARMv6        Distribution = "linux/arm/v6"
	DistLinuxARMv7        Distribution = "linux/arm/v7"
	DistLinuxARM64        Distribution = "linux/arm64"
	DistLinuxLoong64      Distribution = "linux/loong64"
	DistLinuxMips         Distribution = "linux/mips"
	DistLinuxMips64       Distribution = "linux/mips64"
	DistLinuxMips64le     Distribution = "linux/mips64le"
	DistLinuxMipsle       Distribution = "linux/mipsle"
	DistLinuxPPC64        Distribution = "linux/ppc64"
	DistLinuxPPC64le      Distribution = "linux/ppc64le"
	DistLinuxRISCV64      Distribution = "linux/riscv64"
	DistLinuxS390X        Distribution = "linux/s390x"
)

const (
	DistOpenBSD386          Distribution = "openbsd/386"
	DistOpenBSD386SSE2      Distribution = "openbsd/386/sse2"
	DistOpenBSD386SoftFLoat Distribution = "openbsd/386/softfloat"
	DistOpenBSDAMD64        Distribution = "openbsd/amd64"
	DistOpenBSDAMD64v1      Distribution = "openbsd/amd64/v1"
	DistOpenBSDAMD64v2      Distribution = "openbsd/amd64/v2"
	DistOpenBSDAMD64v3      Distribution = "openbsd/amd64/v3"
	DistOpenBSDAMD64v4      Distribution = "openbsd/amd64/v4"
	DistOpenBSDARM          Distribution = "openbsd/arm"
	DistOpenBSDARMv6        Distribution = "openbsd/arm/v6"
	DistOpenBSDARMv7        Distribution = "openbsd/arm/v7"
	DistOpenBSDARM64        Distribution = "openbsd/arm64"
	DistOpenBSDMips64       Distribution = "openbsd/mips64"
)

const (
	DistPlan9386          Distribution = "plan9/386"
	DistPlan9386SSE2      Distribution = "plan9/386/sse2"
	DistPlan9386SoftFloat Distribution = "plan9/386/softfloat"
	DistPlan9AMD64        Distribution = "plan9/amd64"
	DistPlan9AMD64v1      Distribution = "plan9/amd64/v1"
	DistPlan9AMD64v2      Distribution = "plan9/amd64/v2"
	DistPlan9AMD64v3      Distribution = "plan9/amd64/v3"
	DistPlan9AMD64v4      Distribution = "plan9/amd64/v4"
	DistPlan9ARM          Distribution = "plan9/arm/v6"
	DistPlan9ARMv6        Distribution = "plan9/arm/v6"
	DistPlan9ARMv7        Distribution = "plan9/arm/v7"
)

const (
	DistSolarisAMD64   Distribution = "solaris/amd64"
	DistSolarisAMD64v1 Distribution = "solaris/amd64/v1"
	DistSolarisAMD64v2 Distribution = "solaris/amd64/v2"
	DistSolarisAMD64v3 Distribution = "solaris/amd64/v3"
	DistSolarisAMD64v4 Distribution = "solaris/amd64/v4"
)

const (
	DistWindows386          Distribution = "windows/386"
	DistWindows386SSE2      Distribution = "windows/386/sse2"
	DistWindows386SoftFloat Distribution = "windows/386/softfloat"
	DistWindowsAMD64        Distribution = "windows/amd64"
	DistWindowsAMD64v1      Distribution = "windows/amd64/v1"
	DistWindowsAMD64v2      Distribution = "windows/amd64/v2"
	DistWindowsAMD64v3      Distribution = "windows/amd64/v3"
	DistWindowsAMD64v4      Distribution = "windows/amd64/v4"
	DistWindowsARM          Distribution = "windows/arm"
	DistWindowsARMv6        Distribution = "windows/arm/v6"
	DistWindowsARMv7        Distribution = "windows/arm/v7"
	DistWindowsARM64        Distribution = "windows/arm64"
)

func DistrosFromStringSlice(s []string) []Distribution {
	d := make([]Distribution, len(s))
	for i, v := range s {
		d[i] = Distribution(v)
	}

	return d
}

func DistroOneOf(d Distribution, distros []Distribution) bool {
	for _, v := range distros {
		if d == v {
			return true
		}
	}

	return false
}

func IsWindows(d Distribution) bool {
	return strings.Split(string(d), "/")[0] == "windows"
}

type (
	BuildMode string
	GoARM     string
	GoAMD64   string
	Go386     string
)

const (
	BuildModeDefault BuildMode = "default"
	BuildModeExe     BuildMode = "exe"
)

const (
	GOARM5 GoARM = "5"
	GOARM6 GoARM = "6"
	GOARM7 GoARM = "7"
)

const (
	Go386SSE2      Go386 = "sse2"
	Go386SoftFloat Go386 = "softfloat"
)

type GoBuildOpts struct {
	// Main is the path to the 'main' package that is being compiled.
	Main string

	// Workdir should be the root of the project, ideally where the go.mod lives.
	Workdir string

	// OS is value supplied to the GOOS environment variable
	OS string

	// Arch is value supplied to the GOARCH environment variable
	Arch string

	// BuildMode: The 'go build' and 'go install' commands take a -buildmode argument which
	// indicates which kind of object file is to be built. Currently supported values
	// are visible with the 'go help buildmode' command
	BuildMode BuildMode

	// LDFlags are provided to the '-ldflags' argument. 'arguments to pass on each go tool link invocation.'
	LDFlags map[string][]string

	// ExperimentalFlags are Go build-time feature flags in the "GOEXPERIMENT" environment variable that enable experimental features.
	ExperimentalFlags []string

	// CGOEnabled defines whether or not the CGO_ENABLED flag is set.
	CGOEnabled bool

	// GOARM: For GOARCH=arm, the ARM architecture for which to compile.
	// Valid values are 5, 6, 7.
	GoARM GoARM

	// GO386: For GOARCH=386, how to implement floating point instructions.
	// Valid values are sse2 (default), softfloat.
	Go386 Go386

	// CC is the command to use to compile C code when CGO is enabled. (Sets the "CC" environment variable)
	CC string

	// Output is the path where the compiled artifact should be produced; the '-o' flag basically.
	Output string

	// TrimPath trims filepaths from symbols in the binary, making the binary more reproducible.
	TrimPath bool

	// Tags: a list of additional build tags to consider satisfied
	// during the build. For more information about build tags, see
	// 'go help buildconstraint'.
	Tags []string
}

func OSAndArch(d Distribution) (string, string) {
	p := strings.Split(string(d), "/")
	return p[0], p[1]
}

// GoBuildEnv returns the environment variables that must be set for a 'go build' command given the provided 'GoBuildOpts'.
func GoBuildEnv(opts *GoBuildOpts) map[string]string {
	var (
		os   = opts.OS
		arch = opts.Arch
	)

	env := map[string]string{"GOOS": os, "GOARCH": arch}

	switch arch {
	case "arm":
		env["GOARM"] = string(opts.GoARM)
	}

	if opts.CGOEnabled {
		env["CGO_ENABLED"] = "1"
	} else {
		env["CGO_ENABLED"] = "0"
	}

	if opts.ExperimentalFlags != nil {
		env["GOEXPERIMENT"] = strings.Join(opts.ExperimentalFlags, ",")
	}

	return env
}

func GoLDFlags(opts *GoBuildOpts) string {
	ldflags := strings.Builder{}
	for k, v := range opts.LDFlags {
		if v == nil {
			ldflags.WriteString(k + " ")
			continue
		}

		for _, value := range v {
			// For example, "-X 'main.version=v1.0.0'"
			ldflags.WriteString(fmt.Sprintf("%s '%s' ", k, value))
		}
	}

	return ldflags.String()
}

// GoBuildCmd returns the exec.Cmd that best represents the 'go build' command given the options provided in 'GoBuildOpts'.
// Note that some incompatible arguments, like using a non-ARM distribution along with the 'arm' argument will be avoided by the function.
func GoBuildCmd(opts *GoBuildOpts) *exec.Cmd {
	args := []string{"build"}

	if opts.LDFlags != nil {
		args = append(args, "-ldflags", GoLDFlags(opts))
	}

	if opts.Output != "" {
		args = append(args, "-o", opts.Output)
	}

	if opts.Tags != nil {
		args = append(args, "-tags", strings.Join(opts.Tags, ","))
	}

	if opts.TrimPath {
		args = append(args, "-trimpath")
	}

	// Go is weird and paths referring to packages within a module to be prefixed with "./".
	// Otherwise, the path is assumed to be relative to $GOROOT
	args = append(args, "./"+opts.Main)

	cmd := exec.Command("go", args...)
	cmd.Path = opts.Workdir
	return cmd
}
