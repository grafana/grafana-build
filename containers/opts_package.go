package containers

import (
	"github.com/grafana/grafana-build/cliutil"
	"github.com/grafana/grafana-build/executil"
)

type PackageOpts struct {
	Distros []executil.Distribution
	Edition string
}

func PackageOptsFromFlags(c cliutil.CLIContext) *PackageOpts {
	d := c.StringSlice("distro")
	distros := make([]executil.Distribution, len(d))
	for i, v := range d {
		distros[i] = executil.Distribution(v)
	}

	return &PackageOpts{
		Distros: distros,
		Edition: c.String("edition"),
	}
}
