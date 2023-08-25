package executil

import (
	"fmt"
	"os/exec"
	"strings"
)

type (
	BuildMode string
	GoARM     string
	GoAMD64   string
	Go386     string
	LibC      int
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

const (
	Musl LibC = iota
	GLibC
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

	// CXX is the command to use to compile C++ code when CGO is enabled. (Sets the "CXX" environment variable)
	CXX string

	// Output is the path where the compiled artifact should be produced; the '-o' flag basically.
	Output string

	// TrimPath trims filepaths from symbols in the binary, making the binary more reproducible.
	TrimPath bool

	// Tags: a list of additional build tags to consider satisfied
	// during the build. For more information about build tags, see
	// 'go help buildconstraint'.
	Tags []string

	// LibC doesn't change how 'go build' is ran, but it does affect what base image is used.
	LibC LibC
}

// GoBuildEnv returns the environment variables that must be set for a 'go build' command given the provided 'GoBuildOpts'.
func GoBuildEnv(opts *GoBuildOpts) map[string]string {
	var (
		os   = opts.OS
		arch = opts.Arch
	)

	env := map[string]string{"GOOS": os, "GOARCH": arch}

	if arch == "arm" {
		env["GOARM"] = string(opts.GoARM)
	}

	if opts.CGOEnabled {
		env["CGO_ENABLED"] = "1"

		// https://github.com/mattn/go-sqlite3/issues/1164#issuecomment-1635253695
		env["CGO_CFLAGS"] = "-D_LARGEFILE64_SOURCE"
	} else {
		env["CGO_ENABLED"] = "0"
	}

	if opts.ExperimentalFlags != nil {
		env["GOEXPERIMENT"] = strings.Join(opts.ExperimentalFlags, ",")
	}

	if opts.CC != "" {
		env["CC"] = opts.CC
	}

	if opts.CXX != "" {
		env["CXX"] = opts.CXX
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
