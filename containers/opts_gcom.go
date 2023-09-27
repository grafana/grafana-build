package containers

import "github.com/grafana/grafana-build/cliutil"

// GCOMOpts are options used when making requests to grafana.com.
type GCOMOpts struct {
	URL         string
	ApiKey      string
	DownloadURL string
}

func GCOMOptsFromFlags(c cliutil.CLIContext) *GCOMOpts {
	return &GCOMOpts{
		URL:         c.String("url"),
		ApiKey:      c.String("api-key"),
		DownloadURL: c.String("download-url"),
	}
}
