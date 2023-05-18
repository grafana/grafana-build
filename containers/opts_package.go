package containers

import (
	"dagger.io/dagger"
	"github.com/grafana/grafana-build/cliutil"
	"github.com/grafana/grafana-build/executil"
)

type PackageOpts struct {
	Distros  []executil.Distribution
	Platform dagger.Platform
	Edition  string
}

func PackageOptsFromFlags(c cliutil.CLIContext) *PackageOpts {
	d := c.StringSlice("distro")
	distros := make([]executil.Distribution, len(d))
	for i, v := range d {
		distros[i] = executil.Distribution(v)
	}

	return &PackageOpts{
		Distros:  distros,
		Platform: dagger.Platform(c.String("platform")),
		Edition:  c.String("edition"),
	}
}
