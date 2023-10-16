package containers

import "github.com/grafana/grafana-build/cliutil"

type ValidateOpts struct {
	Type   string
	Distro string
}

func ValidateOptsFromFlags(c cliutil.CLIContext) *ValidateOpts {
	return &ValidateOpts{
		Type:   c.String("type"),
		Distro: c.String("distro"),
	}
}
