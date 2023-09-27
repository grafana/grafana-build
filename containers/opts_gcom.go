package containers

import "github.com/grafana/grafana-build/cliutil"

// GCOMOpts are options used when making requests to grafana.com.
type GCOMOpts struct {
	GCOMEnabled bool
	GCOMApiKey  string
	GCOMUrl     string
}

func GCOMOptsFromFlags(c cliutil.CLIContext) *GCOMOpts {
	return &GCOMOpts{
		GCOMEnabled: c.Bool("gcom"),
		GCOMApiKey:  c.String("gcom-api-key"),
		GCOMUrl:     c.String("gcom-url"),
	}
}
