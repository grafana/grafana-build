package containers

import (
	"net/url"

	"github.com/grafana/grafana-build/cliutil"
)

// GCOMOpts are options used when making requests to grafana.com.
type GCOMOpts struct {
	URL         *url.URL
	DownloadURL *url.URL
	ApiKey      string
}

func GCOMOptsFromFlags(c cliutil.CLIContext) (*GCOMOpts, error) {
	apiUrl, err := url.Parse(c.String("url"))
	if err != nil {
		return nil, err
	}
	downloadUrl, err := url.Parse(c.String("download-url"))
	if err != nil {
		return nil, err
	}
	return &GCOMOpts{
		URL:         apiUrl,
		DownloadURL: downloadUrl,
		ApiKey:      c.String("api-key"),
	}, nil
}
