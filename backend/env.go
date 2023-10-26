package backend

import "strings"

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
	// OS is value supplied to the GOOS environment variable
	OS string

	// Arch is value supplied to the GOARCH environment variable
	Arch string

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

// ViceroyEnv returns the environment variables that must be set for a 'go build' command given the provided 'GoBuildOpts'.
func ViceroyEnv(opts *GoBuildOpts) map[string]string {
	var (
		os   = opts.OS
		arch = opts.Arch
	)

	env := map[string]string{"VICEROYOS": os, "VICEROYARCH": arch}

	if arch == "arm" {
		env["VICEROYARM"] = string(opts.GoARM)
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
		env["CC"] = "viceroycc"
	}

	return env
}
