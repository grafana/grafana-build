package artifacts

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/cliutil"
	"github.com/grafana/grafana-build/cmd/flags"
	"github.com/grafana/grafana-build/daggerutil"
	"github.com/grafana/grafana-build/git"
	"github.com/grafana/grafana-build/pipeline"
)

// GrafnaaOpts are populated by the 'GrafanaFlags' flags.
// These options define how to mount or clone the grafana/enterprise source code.
type GrafanaDirectoryOpts struct {
	Enabled bool
	// GrafanaDir is the path to the Grafana source tree.
	// If GrafanaDir is empty, then we're most likely cloning Grafana and using that as a directory.
	GrafanaDir string
	// GrafanaRepo will clone Grafana from a different repository when cloning Grafana.
	GrafanaRepo string
	// GrafanaRef will checkout a specific tag, branch, or commit when cloning Grafana.
	GrafanaRef string

	// GitHubToken is used when cloning Grafana/Grafana Enterprise.
	GitHubToken string
}

func GrafanaDirectoryOptsFromFlags(ctx context.Context, c cliutil.CLIContext) (*GrafanaDirectoryOpts, error) {
	return &GrafanaDirectoryOpts{
		Enabled:     c.Bool("enabled"),
		GrafanaRepo: c.String("grafana-repo"),
		GrafanaDir:  c.String("grafana-dir"),
		GrafanaRef:  c.String("grafana-ref"),
		GitHubToken: c.String("github-token"),
	}, nil
}

func grafanaDirectory(ctx context.Context, c cliutil.CLIContext, d *dagger.Client) (any, error) {
	opts, err := GrafanaDirectoryOptsFromFlags(ctx, c)
	if err != nil {
		return nil, err
	}

	// If GrafanaDir was provided, then we can just use that one.
	if path := opts.GrafanaDir; path != "" {
		slog.Info("Using local Grafana found", "path", path)
		src, err := daggerutil.HostDir(d, path)
		if err != nil {
			return nil, err
		}

		return src, nil
	}

	// Since GrafanaDir was not provided, we must clone it.
	ght := opts.GitHubToken

	// If GitHubToken was not set from flag
	if ght == "" {
		log.Println("Looking up github token from 'GITHUB_TOKEN' environment variable or '$XDG_HOME/.gh'")
		token, err := git.LookupGitHubToken(ctx)
		if err != nil {
			return nil, err
		}
		if token == "" {
			return nil, fmt.Errorf("unable to acquire github token")
		}
		ght = token
	}

	src, err := git.CloneWithGitHubToken(d, ght, opts.GrafanaRepo, opts.GrafanaRef)
	if err != nil {
		return nil, err
	}

	return src, nil
}

// ArgumentGrafanaDirectory will provide the valueFunc that initializes and returns a *dagger.Directory that has Grafana in it.
// This Grafana directory could be initialized with Grafana Enterprise if the appropriate options are provided.
// It could also use a local directory instead of cloning it.
// Where possible, when cloning and no authentication options are provided, the valuefunc will try to use the configured github CLI for cloning.
var ArgumentGrafanaDirectory = pipeline.Argument{
	Name:        "grafana-dir",
	Description: "The grafana backend binaries ('grafana', 'grafana-cli', 'grafana-server') in a directory",
	Flags:       flags.Grafana,
	ValueFunc:   grafanaDirectory,
	Required:    true,
}
