package arguments

import (
	"context"

	"github.com/grafana/grafana-build/cliutil"
	"github.com/grafana/grafana-build/pipeline"
)

type EnterpriseDirectoryOpts struct {
	Enabled bool
	// EnterpriseDir is the path to the Grafana Enterprise source tree if it exists locally.
	// If EnterpriseDir is empty, then we're most likely cloning Grafana Enterprise and using that as a directory.
	EnterpriseDir string
	// EnterpriseRepo will clone Grafana Enterprise from a different repository.
	EnterpriseRepo string
	// EnterpriseRef will checkout a specific tag, branch, or commit with cloning Grafana Enterprise.
	EnterpriseRef string

	GitHubToken string
}

func EnterpriseDirectoryOptsFromFlags(ctx context.Context, c cliutil.CLIContext) (*EnterpriseDirectoryOpts, error) {
	return &EnterpriseDirectoryOpts{
		Enabled:        c.Bool("enterprise"),
		EnterpriseRepo: c.String("enterprise-repo"),
		EnterpriseDir:  c.String("enterprise-dir"),
		EnterpriseRef:  c.String("enterprise-ref"),
		GitHubToken:    c.String("github-token"),
	}, nil
}

var EnterpriseDirectory = pipeline.Argument{
	Name: "enterprise-dir",
}
