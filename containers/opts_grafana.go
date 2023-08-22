package containers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/cliutil"
	"github.com/grafana/grafana-build/stringutil"
)

// GrafnaaOpts are populated by the 'GrafanaFlags' flags.
// These options define how to mount or clone the grafana/enterprise source code.
type GrafanaOpts struct {
	BuildGrafana bool
	// GrafanaDir is the path to the Grafana source tree.
	GrafanaDir      string
	GrafanaRepo     string
	GrafanaRef      string
	BuildEnterprise bool
	EnterpriseRepo  string
	EnterpriseRef   string
	// EnterpriseDir is the path to the Grafana Enterprise source tree.
	EnterpriseDir string
	BuildID       string
	GitHubToken   string
	Env           map[string]string
	GoTags        []string
	GoVersion     string

	// Version will be set by the '--version' flag if provided, and returned in the 'Version' function.
	// If not set, then the version function will attempt to retrieve the version from Grafana's package.json or some other method.
	Version          string
	YarnCacheHostDir string
}

func GrafanaOptsFromFlags(ctx context.Context, c cliutil.CLIContext) (*GrafanaOpts, error) {
	var (
		version        = c.String("version")
		grafana        = c.Bool("grafana")
		grafanaRepo    = c.String("grafana-repo")
		grafanaDir     = c.String("grafana-dir")
		ref            = c.String("grafana-ref")
		enterprise     = c.Bool("enterprise")
		enterpriseDir  = c.String("enterprise-dir")
		enterpriseRepo = c.String("enterprise-repo")
		enterpriseRef  = c.String("enterprise-ref")
		buildID        = c.String("build-id")
		gitHubToken    = c.String("github-token")
		goTags         = c.StringSlice("go-tags")
		goVersion      = c.String("go-version")
	)

	if buildID == "" {
		buildID = stringutil.RandomString(12)
	}

	// If the user has provided any ref except the default, then
	// we can safely assume they want to compile enterprise.
	// If they've explicitly set the enterprise flag to emptystring then we can assume they want it to be false.
	if enterpriseRef == "" {
		enterprise = false
	} else if enterpriseRef != "main" {
		enterprise = true
	}

	// If the user has set the Enterprise Directory, then
	// we can safely assume that they want to compile enterprise.
	if enterpriseDir != "" {
		if _, err := os.Stat(enterpriseDir); err != nil {
			return nil, fmt.Errorf("stat enterprise dir '%s': %w", enterpriseDir, err)
		}
		enterprise = true
	}
	env := map[string]string{}
	for _, v := range c.StringSlice("env") {
		if p := strings.Split(v, "="); len(p) > 1 {
			key := p[0]
			value := strings.Join(p[1:], "=")
			env[key] = value
		}
	}

	return &GrafanaOpts{
		BuildID:          buildID,
		Version:          version,
		BuildEnterprise:  enterprise,
		BuildGrafana:     grafana,
		GrafanaDir:       grafanaDir,
		GrafanaRepo:      grafanaRepo,
		GrafanaRef:       ref,
		EnterpriseDir:    enterpriseDir,
		EnterpriseRepo:   enterpriseRepo,
		EnterpriseRef:    enterpriseRef,
		GitHubToken:      gitHubToken,
		Env:              env,
		GoTags:           goTags,
		GoVersion:        goVersion,
		YarnCacheHostDir: c.String("yarn-cache"),
	}, nil
}

func (g *GrafanaOpts) DetectVersion(ctx context.Context, client *dagger.Client, grafanaDir *dagger.Directory) (string, error) {
	// If Version is set, we should use that.
	// If it's not set, then use the GetPackageJSONVersion function to get the version and return that
	if g.Version != "" {
		return g.Version, nil
	}

	log.Println("Version not provided; getting version from package.json...")
	v, err := GetJSONValue(ctx, client, grafanaDir, "package.json", "version")
	if err != nil {
		return "", err
	}

	log.Println("Got version", v)
	return v, nil
}

// Entrerprise will attempt to mount or clone Grafana Enterprise based on the arguments provided. If, for example, the enterprise-dir argument was supplied, then this function will mount that directory.
// If it was not set then it will attempt to clone, optionally using the 'enterprise-ref' argument.
func (g *GrafanaOpts) Enterprise(ctx context.Context, grafana *dagger.Directory, client *dagger.Client) (*dagger.Directory, error) {
	// If GrafanaDir was provided, then we can just use that one.
	if path := g.EnterpriseDir; path != "" {
		src, err := HostDir(client, path)
		if err != nil {
			return nil, err
		}

		return InitializeEnterprise(client, grafana, src), nil
	}

	// Since GrafanaDir was not provided, we must clone it.
	ght := g.GitHubToken

	// If GitHubToken was not set from flag
	if ght == "" {
		log.Println("Looking up github token from 'GITHUB_TOKEN' environment variable or '$XDG_HOME/.gh'")
		token, err := LookupGitHubToken(ctx)
		if err != nil {
			return nil, err
		}
		if token == "" {
			return nil, fmt.Errorf("unable to acquire github token")
		}
		ght = token
	}

	log.Printf("Cloning Grafana Enterprise...")
	src, err := CloneWithGitHubToken(client, ght, g.EnterpriseRepo, g.EnterpriseRef)
	if err != nil {
		return nil, err
	}

	return InitializeEnterprise(client, grafana, src), nil
}

// Grafana will attempt to mount or clone Grafana based on the arguments provided. If, for example, the grafana-dir argument was supplied, then this function will mount that directory.
// If it was not set then it will attempt to clone, optionally using the 'grafana-ref' argument.
func (g *GrafanaOpts) Grafana(ctx context.Context, client *dagger.Client) (*dagger.Directory, error) {
	// If GrafanaDir was provided, then we can just use that one.
	if path := g.GrafanaDir; path != "" {
		log.Println("Using local Grafana found at", path)
		src, err := HostDir(client, path)
		if err != nil {
			return nil, err
		}

		return src, nil
	}

	// Since GrafanaDir was not provided, we must clone it.
	ght := g.GitHubToken

	// If GitHubToken was not set from flag
	if ght == "" {
		log.Println("Looking up github token from 'GITHUB_TOKEN' environment variable or '$XDG_HOME/.gh'")
		token, err := LookupGitHubToken(ctx)
		if err != nil {
			return nil, err
		}
		if token == "" {
			return nil, fmt.Errorf("unable to acquire github token")
		}
		ght = token
	}

	log.Printf("Cloning Grafana...")
	src, err := CloneWithGitHubToken(client, ght, g.GrafanaRepo, g.GrafanaRef)
	if err != nil {
		return nil, err
	}

	return src, nil
}
