package arguments

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/grafana/grafana-build/cliutil"
	"github.com/grafana/grafana-build/daggerutil"
	"github.com/grafana/grafana-build/git"
	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
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

func grafanaDirectory(ctx context.Context, opts *pipeline.ArgumentOpts) (any, error) {
	o, err := GrafanaDirectoryOptsFromFlags(ctx, opts.CLIContext)
	if err != nil {
		return nil, err
	}

	// If GrafanaDir was provided, then we can just use that one.
	if path := o.GrafanaDir; path != "" {
		slog.Info("Using local Grafana found", "path", path)
		src, err := daggerutil.HostDir(opts.Client, path)
		if err != nil {
			return nil, err
		}

		return src, nil
	}

	// Since GrafanaDir was not provided, we must clone it.
	ght := o.GitHubToken

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

	src, err := git.CloneWithGitHubToken(opts.Client, ght, o.GrafanaRepo, o.GrafanaRef)
	if err != nil {
		return nil, err
	}

	return src, nil
}

var GrafanaDirectoryFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "grafana-dir",
		Usage:    "Local Grafana dir to use, instead of git clone",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "grafana-repo",
		Usage:    "Grafana repo to clone, not valid if --grafana-dir is set",
		Required: false,
		Value:    "https://github.com/grafana/grafana.git",
	},
	&cli.StringFlag{
		Name:     "grafana-ref",
		Usage:    "Grafana ref to clone, not valid if --grafana-dir is set",
		Required: false,
		Value:    "main",
	},
	&cli.StringFlag{
		Name:     "github-token",
		Usage:    "Github token to use for git cloning, by default will be pulled from GitHub",
		Required: false,
	},
}

// GrafanaDirectory will provide the valueFunc that initializes and returns a *dagger.Directory that has Grafana in it.
// This Grafana directory could be initialized with Grafana Enterprise if the appropriate options are provided.
// It could also use a local directory instead of cloning it.
// Where possible, when cloning and no authentication options are provided, the valuefunc will try to use the configured github CLI for cloning.
var GrafanaDirectory = pipeline.Argument{
	Name:        "grafana-dir",
	Description: "The grafana backend binaries ('grafana', 'grafana-cli', 'grafana-server') in a directory",
	Flags:       GrafanaDirectoryFlags,
	ValueFunc:   grafanaDirectory,
}
