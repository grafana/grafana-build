package containers

import "github.com/grafana/grafana-build/cliutil"

type ProImageOpts struct {
	// The path to a Grafana debian package.
	Deb string

	// The Grafana version.
	GrafanaVersion string

	// The release type.
	ReleaseType string

	// True if the pro image should be pushed to the container registry.
	Push bool

	// Github token used to clone private repositories.
	GithubToken string
}

func ProImageOptsFromFlags(c cliutil.CLIContext) *ProImageOpts {
	return &ProImageOpts{
		GithubToken:    c.String("github-token"),
		Deb:            c.String("deb"),
		GrafanaVersion: c.String("grafana-version"),
		ReleaseType:    c.String("release-type"),
		Push:           c.Bool("push"),
	}
}
