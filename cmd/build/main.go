// package main defines the Go program that uses Dagger to clone and build a Grafana.
package main

import (
	"context"
	"os"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
	flag "github.com/spf13/pflag"
)

type Flags struct {
	Path    string
	Version string
	Verbose bool
}

func ParseFlags(args []string) (*Flags, error) {
	var (
		version string
		verbose bool
		path    string
	)

	f := flag.NewFlagSet("build", flag.ExitOnError)
	f.StringVar(&version, "version", "v0.0.0", "The Grafana version being built. This value is injected into a const in the binary and is used to render the Grafana version in the UI. When running in CI, this flag should be set to the 'DRONE_TAG' environment variable or equivalent.")
	f.BoolVarP(&verbose, "verbose", "v", false, "Verbose logging enabled")
	f.StringVar(&path, "path", ".grafana", "The path to the directory where Grafana is cloned. This directory must be a git repository. If it does not exist, then Grafana will be cloned and the ref matching 'version' will be checked out.")
	if err := f.Parse(args); err != nil {
		return nil, err
	}

	return &Flags{
		Path:    path,
		Version: version,
		Verbose: verbose,
	}, nil
}

func main() {
	ctx := context.Background()

	args, err := ParseFlags(os.Args)
	if err != nil {
		panic(err)
	}

	opts := []dagger.ClientOpt{}
	if args.Verbose {
		opts = append(opts, dagger.WithLogOutput(os.Stderr))
	}

	d, err := dagger.Connect(ctx, opts...)
	if err != nil {
		panic(err)
	}

	buildinfo, err := containers.GetBuildInfo(ctx, d, args.Path, args.Version)
	if err != nil {
		panic(err)
	}

	distro := executil.DistLinuxAMD64

	if _, err := containers.CompileBackend(d, executil.DistLinuxAMD64, args.Path, buildinfo).Export(ctx, filepath.Join("bin", string(distro))); err != nil {
		panic(err)
	}
}
