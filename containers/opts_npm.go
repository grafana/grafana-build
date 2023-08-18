package containers

import "github.com/grafana/grafana-build/cliutil"

type NPMOpts struct {
	// Registry is the package registry.
	// Uses npmjs by default.
	// Example: registry.npmjs.org
	Registry string

	// Token is supplied to login to the package registry when publishing packages.
	Token string
}

func NPMOptsFromFlags(c cliutil.CLIContext) *NPMOpts {
	return &NPMOpts{
		Registry: c.String("registry"),
		Token:    c.String("token"),
	}
}
