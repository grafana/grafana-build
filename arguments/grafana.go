package arguments

import (
	"context"
	"fmt"
	"log/slog"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/cliutil"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/daggerutil"
	"github.com/grafana/grafana-build/frontend"
	"github.com/grafana/grafana-build/git"
	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
)

const BusyboxImage = "busybox:1.36"

func InitializeEnterprise(d *dagger.Client, grafana *dagger.Directory, enterprise *dagger.Directory) *dagger.Directory {
	hash := d.Container().From("alpine/git").
		WithDirectory("/src/grafana-enterprise", enterprise).
		WithWorkdir("/src/grafana-enterprise").
		WithEntrypoint([]string{}).
		WithExec([]string{"/bin/sh", "-c", "git rev-parse HEAD > .buildinfo.enterprise-commit"}).
		File("/src/grafana-enterprise/.buildinfo.enterprise-commit")

	return d.Container().From(BusyboxImage).
		WithDirectory("/src/grafana", grafana).
		WithDirectory("/src/grafana-enterprise", enterprise).
		WithWorkdir("/src/grafana-enterprise").
		WithFile("/src/grafana/.buildinfo.enterprise-commit", hash).
		WithExec([]string{"/bin/sh", "build.sh"}).
		WithExec([]string{"cp", "LICENSE", "../grafana"}).
		Directory("/src/grafana")
}

// GrafnaaOpts are populated by the 'GrafanaFlags' flags.
// These options define how to mount or clone the grafana/enterprise source code.
type GrafanaDirectoryOpts struct {
	// GrafanaDir is the path to the Grafana source tree.
	// If GrafanaDir is empty, then we're most likely cloning Grafana and using that as a directory.
	GrafanaDir    string
	EnterpriseDir string
	// GrafanaRepo will clone Grafana from a different repository when cloning Grafana.
	GrafanaRepo    string
	EnterpriseRepo string
	// GrafanaRef will checkout a specific tag, branch, or commit when cloning Grafana.
	GrafanaRef    string
	EnterpriseRef string
	// GitHubToken is used when cloning Grafana/Grafana Enterprise.
	GitHubToken string
}

func (o *GrafanaDirectoryOpts) githubToken(ctx context.Context) (string, error) {
	// Since GrafanaDir was not provided, we must clone it.
	ght := o.GitHubToken

	// If GitHubToken was not set from flag
	if ght != "" {
		return ght, nil
	}

	token, err := git.LookupGitHubToken(ctx)
	if err != nil {
		return "", err
	}
	if token == "" {
		return "", fmt.Errorf("unable to acquire github token")
	}

	return token, nil
}

func GrafanaDirectoryOptsFromFlags(ctx context.Context, c cliutil.CLIContext) (*GrafanaDirectoryOpts, error) {
	return &GrafanaDirectoryOpts{
		GrafanaRepo:    c.String("grafana-repo"),
		EnterpriseRepo: c.String("enterprise-repo"),
		GrafanaDir:     c.String("grafana-dir"),
		EnterpriseDir:  c.String("enterprise-dir"),
		GrafanaRef:     c.String("grafana-ref"),
		EnterpriseRef:  c.String("enterprise-ref"),
		GitHubToken:    c.String("github-token"),
	}, nil
}

func cloneOrMount(ctx context.Context, client *dagger.Client, localPath, repo, ref string, o *GrafanaDirectoryOpts) (*dagger.Directory, error) {
	// If GrafanaDir was provided, then we can just use that one.
	if path := localPath; path != "" {
		slog.Info("Using local Grafana found", "path", path)
		return daggerutil.HostDir(client, path)
	}

	ght, err := o.githubToken(ctx)
	if err != nil {
		return nil, err
	}

	src, err := git.CloneWithGitHubToken(client, ght, repo, ref)
	if err != nil {
		return nil, err
	}

	return src, nil
}

func grafanaDirectory(ctx context.Context, opts *pipeline.ArgumentOpts) (any, error) {
	o, err := GrafanaDirectoryOptsFromFlags(ctx, opts.CLIContext)
	if err != nil {
		return nil, err
	}

	src, err := cloneOrMount(ctx, opts.Client, o.GrafanaDir, o.GrafanaRepo, o.GrafanaRef, o)
	if err != nil {
		return nil, err
	}

	nodeVersion, err := frontend.NodeVersion(opts.Client, src).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get node version from source code: %w", err)
	}

	yarnCache, err := opts.State.CacheVolume(ctx, YarnCacheDirectory)
	if err != nil {
		return nil, err
	}

	container := frontend.YarnInstall(opts.Client, src, nodeVersion, yarnCache, opts.Platform)

	if _, err := containers.ExitError(ctx, container); err != nil {
		return nil, err
	}

	return container.Directory("/src"), nil
}

func enterpriseDirectory(ctx context.Context, opts *pipeline.ArgumentOpts) (any, error) {
	// Get the Grafana directory...
	o, err := GrafanaDirectoryOptsFromFlags(ctx, opts.CLIContext)
	if err != nil {
		return nil, err
	}

	grafanaDir, err := grafanaDirectory(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error initializing grafana directory: %w", err)
	}

	src, err := cloneOrMount(ctx, opts.Client, o.EnterpriseDir, o.EnterpriseRepo, o.EnterpriseRef, o)
	if err != nil {
		return nil, err
	}

	return InitializeEnterprise(opts.Client, grafanaDir.(*dagger.Directory), src), nil
}

var GrafanaDirectoryFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "grafana-dir",
		Usage:    "Local Grafana dir to use, instead of git clone",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "enterprise-dir",
		Usage:    "Local Grafana Enterprise dir to use, instead of git clone",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "grafana-repo",
		Usage:    "Grafana repo to clone, not valid if --grafana-dir is set",
		Required: false,
		Value:    "https://github.com/grafana/grafana.git",
	},
	&cli.StringFlag{
		Name:     "enterprise-repo",
		Usage:    "Grafana Enterprise repo to clone, not valid if --grafana-dir is set",
		Required: false,
		Value:    "https://github.com/grafana/grafana-enterprise.git",
	},
	&cli.StringFlag{
		Name:     "grafana-ref",
		Usage:    "Grafana ref to clone, not valid if --grafana-dir is set",
		Required: false,
		Value:    "main",
	},
	&cli.StringFlag{
		Name:     "enterprise-ref",
		Usage:    "Grafana Enterprise ref to clone, not valid if --grafana-dir is set",
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
// Where possible, when cloning and no authentication options are provided, the valuefunc will try to use the configured github CLI for cloning.
var GrafanaDirectory = pipeline.Argument{
	Name:        "grafana-dir",
	Description: "The source tree of the Grafana repository",
	Flags:       GrafanaDirectoryFlags,
	ValueFunc:   grafanaDirectory,
}

// EnterpriseDirectory will provide the valueFunc that initializes and returns a *dagger.Directory that has Grafana Enterprise initialized it.
// Where possible, when cloning and no authentication options are provided, the valuefunc will try to use the configured github CLI for cloning.
var EnterpriseDirectory = pipeline.Argument{
	Name:        "enterprise-dir",
	Description: "The source tree of Grafana Enterprise",
	Flags:       GrafanaDirectoryFlags,
	ValueFunc:   enterpriseDirectory,
}
