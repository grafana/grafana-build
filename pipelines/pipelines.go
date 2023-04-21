// package pipelines has functions and types that orchestrate containers.
package pipelines

import (
	"context"
	"fmt"
	"log"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

type CLIContext interface {
	Bool(string) bool
	String(string) string
	Set(string, string) error
	StringSlice(string) []string
	Path(string) string
}

type PipelineArgs struct {
	// These arguments are ones that are available at the global level.
	Verbose      bool
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

	// Context is available for all sub-commands that define their own flags.
	Context CLIContext

	// Version will be set by the '--version' flag if provided, and returned in the 'Version' function.
	// If not set, then the version function will attempt to retrieve the version from Grafana's package.json or some other method.
	Version string
}

// PipelineArgsFromContext populates a pipelines.PipelineArgs from a CLI context.
func PipelineArgsFromContext(ctx context.Context, c CLIContext) (PipelineArgs, error) {
	var (
		verbose       = c.Bool("v")
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
		buildID = randomString(12)
	}

	// If the user has provided any ref except the default, then
	// we can safely assume they want to compile enterprise.
	// If they've explicitely set the enterprise flag to emptystring then we can assume they want it to be false.
	if enterpriseRef == "" {
		enterprise = false
	} else if enterpriseRef != "main" {
		enterprise = true
	}

	// If the user has set the Enterprise Directory, then
	// we can safely assume that they want to compile enterprise.
	if enterpriseDir != "" {
		if _, err := os.Stat(enterpriseDir); err != nil {
			return PipelineArgs{}, fmt.Errorf("stat enterprise dir '%s': %w", enterpriseDir, err)
		}
		enterprise = true
	}

	return PipelineArgs{
		BuildID:         buildID,
		Verbose:         verbose,
		Version:         version,
		BuildEnterprise: enterprise,
		BuildGrafana:    grafana,
		GrafanaDir:      grafanaDir,
		GrafanaRef:      ref,
		EnterpriseDir:   enterpriseDir,
		EnterpriseRef:   enterpriseRef,
		Context:         c,
		GitHubToken:     gitHubToken,
	}, nil
}

type PipelineFunc func(context.Context, *dagger.Client, *dagger.Directory, PipelineArgs) error

func (p *PipelineArgs) DetectVersion(ctx context.Context, client *dagger.Client, grafanaDir *dagger.Directory) (string, error) {
	// If Version is set, we should use that.
	// If it's not set, then use the containers.GetPackageJSONVersion function to get the version and return that

	if p.Version != "" {
		return p.Version, nil
	}

	log.Println("Version not provided; getting version from package.json...")
	v, err := containers.GetPackageJSONVersion(ctx, client, grafanaDir)
	if err != nil {
		return "", err
	}

	log.Println("Got version", v)
	return v, nil
}

func (p *PipelineArgs) Grafana(ctx context.Context, client *dagger.Client) (*dagger.Directory, error) {
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
	if p.GrafanaDir == "" {
		// Grafana will clone from git, needs auth
		cloneGrafana = true
	}

	if p.BuildEnterprise && p.EnterpriseDir == "" {
		// Enterprise will clone from git, needs auth
		cloneEnterprise = true
	}

	ght := p.GitHubToken

	if cloneEnterprise && p.GitHubToken == "" {
		// If GitHubToken was not set from flag
		log.Println("Aquiring github token")
		token, err := LookupGitHubToken(ctx)
		if err != nil {
			return nil, err
		}
		if token == "" {
			return nil, fmt.Errorf("unable to aquire github token")
		}
		ght = token
	}

	var (
		src *dagger.Directory
	)

	if !cloneGrafana {
		log.Printf("Mounting local directory %s", p.GrafanaDir)
		grafanaSrc, err := containers.MountLocalDir(client, p.GrafanaDir)
		if err != nil {
			return nil, err
		}

		src = grafanaSrc
	} else {
		log.Printf("Cloning Grafana repo from https://github.com/grafana/grafana.git, ref %s", p.GrafanaRef)
		grafanaSrc, err := containers.Clone(client, "https://github.com/grafana/grafana.git", p.GrafanaRef)
		if err != nil {
			return nil, err
		}
		src = grafanaSrc
	}

	// If the enterprise global flag is set, then clone and initialize Grafana Enterprise as well.
	if !p.BuildEnterprise {
		return src, nil
	}

	log.Println("We will build the enterprise version of Grafana")
	var (
		enterpriseDir *dagger.Directory
	)

	if p.EnterpriseDir != "" {
		log.Printf("Mounting local enterprise directory %s", p.EnterpriseDir)
		enterpriseSrcDir, err := containers.MountLocalDir(client, p.EnterpriseDir)
		if err != nil {
			return nil, err
		}

		enterpriseDir = enterpriseSrcDir
	} else {
		log.Printf("Cloning Grafana Enterprise repo from https://github.com/grafana/grafana-enterprise.git, ref %s", p.EnterpriseRef)
		enterpriseSrcDir, err := containers.CloneWithGitHubToken(client, ght, "https://github.com/grafana/grafana-enterprise.git", p.EnterpriseRef)
		if err != nil {
			return nil, err
		}

		enterpriseDir = enterpriseSrcDir
	}

	return containers.InitializeEnterprise(client, src, enterpriseDir), nil
}
