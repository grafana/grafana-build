package containers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	GrafanaRef      string
	BuildEnterprise bool
	EnterpriseRef   string
	// EnterpriseDir is the path to the Grafana Enterprise source tree.
	EnterpriseDir string
	BuildID       string
	GitHubToken   string

	// Version will be set by the '--version' flag if provided, and returned in the 'Version' function.
	// If not set, then the version function will attempt to retrieve the version from Grafana's package.json or some other method.
	Version string
}

func GrafanaOptsFromFlags(ctx context.Context, c cliutil.CLIContext) (*GrafanaOpts, error) {
	var (
		version       = c.String("version")
		grafana       = c.Bool("grafana")
		grafanaDir    = c.String("grafana-dir")
		ref           = c.String("grafana-ref")
		enterprise    = c.Bool("enterprise")
		enterpriseDir = c.String("enterprise-dir")
		enterpriseRef = c.String("enterprise-ref")
		buildID       = c.String("build-id")
		gitHubToken   = c.String("github-token")
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

	return &GrafanaOpts{
		BuildID:         buildID,
		Version:         version,
		BuildEnterprise: enterprise,
		BuildGrafana:    grafana,
		GrafanaDir:      grafanaDir,
		GrafanaRef:      ref,
		EnterpriseDir:   enterpriseDir,
		EnterpriseRef:   enterpriseRef,
		GitHubToken:     gitHubToken,
	}, nil
}

func (g *GrafanaOpts) DetectVersion(ctx context.Context, client *dagger.Client, grafanaDir *dagger.Directory) (string, error) {
	// If Version is set, we should use that.
	// If it's not set, then use the GetPackageJSONVersion function to get the version and return that
	if g.Version != "" {
		return g.Version, nil
	}

	log.Println("Version not provided; getting version from package.json...")
	v, err := GetPackageJSONVersion(ctx, client, grafanaDir)
	if err != nil {
		return "", err
	}

	log.Println("Got version", v)
	return v, nil
}

// Grafana will attempt to mount or clone Grafana based on the arguments provided. If, for example, the grafana-dir argument was supplied, then this function will mount that directory.
// If it was not set then it will attempt to clone.
func (g *GrafanaOpts) Grafana(ctx context.Context, client *dagger.Client) (*dagger.Directory, error) {
	// 1. Determine whether we should clone Grafana / Enterprise
	// 2. Authenticate with GitHub (if necessary)
	// 3. Get the Grafana source tree either locally or from a container where it was cloned
	// 4. Do the same for Enterprise (if necessary) and `build.sh`.
	// 5. Return the directory

	// Determine if we need to git authenticate, and then do it
	var (
		cloneGrafana    bool
		cloneEnterprise bool
	)

	// If GrafanaDir was not provided, then it will need to be cloned.
	if g.GrafanaDir == "" {
		// Grafana will clone from git, needs auth
		cloneGrafana = true
	}

	if g.BuildEnterprise && g.EnterpriseDir == "" {
		// Enterprise will clone from git, needs auth
		cloneEnterprise = true
	}

	ght := g.GitHubToken

	if cloneEnterprise && g.GitHubToken == "" {
		// If GitHubToken was not set from flag
		log.Println("Acquiring github token")
		token, err := LookupGitHubToken(ctx)
		if err != nil {
			return nil, err
		}
		if token == "" {
			return nil, fmt.Errorf("unable to acquire github token")
		}
		ght = token
	}

	var (
		src *dagger.Directory
	)

	if !cloneGrafana {
		log.Printf("Mounting local directory %s", g.GrafanaDir)
		grafanaSrc, err := MountLocalDir(client, g.GrafanaDir)
		if err != nil {
			return nil, err
		}

		src = grafanaSrc
	} else {
		log.Printf("Cloning Grafana repo from https://github.com/grafana/grafana.git, ref %s", g.GrafanaRef)
		grafanaSrc, err := Clone(client, "https://github.com/grafana/grafana.git", g.GrafanaRef)
		if err != nil {
			return nil, err
		}
		src = grafanaSrc
	}

	// If the enterprise global flag is set, then clone and initialize Grafana Enterprise as well.
	if !g.BuildEnterprise {
		return src, nil
	}

	log.Println("We will build the enterprise version of Grafana")
	var (
		enterpriseDir *dagger.Directory
	)

	if g.EnterpriseDir != "" {
		log.Printf("Mounting local enterprise directory %s", g.EnterpriseDir)
		enterpriseSrcDir, err := MountLocalDir(client, g.EnterpriseDir)
		if err != nil {
			return nil, err
		}

		enterpriseDir = enterpriseSrcDir
	} else {
		log.Printf("Cloning Grafana Enterprise repo from https://github.com/grafana/grafana-enterprise.git, ref %s", g.EnterpriseRef)
		enterpriseSrcDir, err := CloneWithGitHubToken(client, ght, "https://github.com/grafana/grafana-enterprise.git", g.EnterpriseRef)
		if err != nil {
			return nil, err
		}

		enterpriseDir = enterpriseSrcDir
	}

	return InitializeEnterprise(client, src, enterpriseDir), nil
}

// LookupGitHubToken will try to find a GitHub access token that can then be used for various API calls but also cloning of private repositories.
func LookupGitHubToken(ctx context.Context) (string, error) {
	log.Print("Looking for a GitHub token")

	// First try: Check if it's in the environment. This can override everything!
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		log.Print("Using GitHub token provided via environment variable")
		return token, nil
	}

	// Next, check if the user has gh installed and *it* has a token set:
	var data bytes.Buffer
	var errData bytes.Buffer
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		return "", fmt.Errorf("GitHub CLI not installed (expected a --github-token flag, a GITHUB_TOKEN environment variable, or a configured GitHub CLI)")
	}

	//nolint:gosec
	cmd := exec.CommandContext(ctx, ghPath, "auth", "token")
	cmd.Stdout = &data
	cmd.Stderr = &errData

	if err := cmd.Run(); err != nil {
		log.Printf("Querying gh for an access token failed: %s", errData.String())
		return "", fmt.Errorf("lookup in gh failed: %w", err)
	}

	log.Print("Using GitHub token provided via gh")
	return strings.TrimSpace(data.String()), nil
}
