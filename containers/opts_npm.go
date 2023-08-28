package containers

import "github.com/grafana/grafana-build/cliutil"

type NPMOpts struct {
	// Registry is the package registry.
	// Uses npmjs by default.
	// Example: registry.npmjs.org
	Registry string

	// Token is supplied to login to the package registry when publishing packages.
	Token string

	// Latest is used to tag the package as latest after published.
	Latest bool

	// Next is used to tag the package as next after published.
	Next bool
}

func NPMOptsFromFlags(c cliutil.CLIContext) *NPMOpts {
	return &NPMOpts{
		Registry: c.String("registry"),
		Token:    c.String("token"),
		Latest:   c.Bool("latest"),
		Next:     c.Bool("next"),
	}
}
