package containers

import "github.com/grafana/grafana-build/cliutil"

type ProImageOpts struct {
	// Github token used to clone private repositories.
	GithubToken string

	// The path to a Grafana debian package.
	Deb string

	// The Grafana version.
	GrafanaVersion string

	// The docker image tag.
	ImageTag string

	// The release type.
	ReleaseType string

	// True if the pro image should be pushed to the container registry.
	Push bool

	// The container registry that the image should be pushed to. Required if Push is true.
	ContainerRegistry string
}

func ProImageOptsFromFlags(c cliutil.CLIContext) *ProImageOpts {
	return &ProImageOpts{
		GithubToken:       c.String("github-token"),
		Deb:               c.String("deb"),
		GrafanaVersion:    c.String("grafana-version"),
		ImageTag:          c.String("image-tag"),
		ReleaseType:       c.String("release-type"),
		Push:              c.Bool("push"),
		ContainerRegistry: c.String("registry"),
	}
}
