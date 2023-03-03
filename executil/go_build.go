package executil

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Distribution is a string that represents the GOOS and GOARCH environment variables joined by a "/".
type Distribution string

// The list of distributions is built from the command "go tool dist list".
// While not all are used, it at least represents the possible combinations.
const (
	DistDarwinAMD64   Distribution = "darwin/amd64"
	DistDarwinARM64   Distribution = "darwin/arm64"
	DistFreeBSD386    Distribution = "freebsd/386"
	DistFreeBSDAMD64  Distribution = "freebsd/amd64"
	DistFreeBSDARM    Distribution = "freebsd/arm"
	DistFreeBSDARM64  Distribution = "freebsd/arm64"
	DistFreeBSDRISCV  Distribution = "freebsd/riscv64"
	DistIllumosAMD64  Distribution = "illumos/amd64"
	DistLinux386      Distribution = "linux/386"
	DistLinuxAMD64    Distribution = "linux/amd64"
	DistLinuxARM      Distribution = "linux/arm"
	DistLinuxARM64    Distribution = "linux/arm64"
	DistLinuxLoong64  Distribution = "linux/loong64"
	DistLinuxMips     Distribution = "linux/mips"
	DistLinuxMips64   Distribution = "linux/mips64"
	DistLinuxMips64le Distribution = "linux/mips64le"
	DistLinuxMipsle   Distribution = "linux/mipsle"
	DistLinuxPPC64    Distribution = "linux/ppc64"
	DistLinuxPPC64le  Distribution = "linux/ppc64le"
	DistLinuxRISCV64  Distribution = "linux/riscv64"
	DistLinuxS390X    Distribution = "linux/s390x"
	DistOpenBSD386    Distribution = "openbsd/386"
	DistOpenBSDAMD64  Distribution = "openbsd/amd64"
	DistOpenBSDARM    Distribution = "openbsd/arm"
	DistOpenBSDARM64  Distribution = "openbsd/arm64"
	DistOpenBSDMips64 Distribution = "openbsd/mips64"
	DistPlan9386      Distribution = "plan9/386"
	DistPlan9AMD64    Distribution = "plan9/amd64"
	DistPlan9ARM      Distribution = "plan9/arm"
	DistSolarisAMD64  Distribution = "solaris/amd64"
	DistWindows386    Distribution = "windows/386"
	DistWindowsAMD64  Distribution = "windows/amd64"
	DistWindowsARM    Distribution = "windows/arm"
	DistWindowsARM64  Distribution = "windows/arm64"
)

func IsWindows(d Distribution) bool {
	return strings.Split(string(d), "/")[0] == "windows"
}

type BuildMode string

const (
	BuildModeDefault BuildMode = "default"
	BuildModeExe     BuildMode = "exe"
)

type GoARM int

const (
	GOARM5 GoARM = 5
	GOARM6 GoARM = 6
	GOARM7 GoARM = 7
)

type Go386 string

const (
	Go386SSE2      Go386 = "sse2"
	Go386SoftFloat Go386 = "softfloat"
)

type GoBuildOpts struct {
	// Main is the path to the 'main' package that is being compiled.
	Main string

	// Workdir should be the root of the project, ideally where the go.mod lives.
	Workdir string

	// Distribution is the combination of OS/architecture that this program is compiled for.
	Distribution Distribution

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

	// CC is the command to use to compile C code when CGO is enabled.
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
	os, arch := OSAndArch(opts.Distribution)

	env := map[string]string{"GOOS": os, "GOARCH": arch}

	switch arch {
	case "arm", "arm64":
		env["GOARM"] = strconv.Itoa(int(opts.GoARM))
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
