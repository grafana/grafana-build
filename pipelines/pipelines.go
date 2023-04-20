// package pipelines has functions and types that orchestrate containers.
package pipelines

import (
	"context"
	"fmt"
	"log"

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

	// ProvidedVersion will be set by the '--version' flag if provided, and returned in the 'Version' function.
	// If not set, then the version function will attempt to retrieve the version from Grafana's package.json or some other method.
	ProvidedVersion string
}

type PipelineFunc func(context.Context, *dagger.Client, PipelineArgs) error

func (p *PipelineArgs) Version(ctx context.Context) (string, error) {
	return p.ProvidedVersion, nil
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
		src             *dagger.Directory
		err             error
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

	if cloneEnterprise && p.GitHubToken == "" {
		// If GitHubToken was not set from flag
		log.Println("Aquiring github token")
		token, err := LookupGitHubToken(ctx)
		if err != nil {
			return nil, err
		}
		if token == "" {
			return nil, fmt.Errorf("Unable to aquire github token")
		}
		p.GitHubToken = token
	}

	// Start by gathering grafana code, either locally or from git
	if cloneGrafana {
		log.Printf("Cloning Grafana repo from https://github.com/grafana/grafana.git, ref %s", p.GrafanaRef)
		src, err = containers.Clone(client, "https://github.com/grafana/grafana.git", p.GrafanaRef)
		if err != nil {
			return nil, err
		}
	} else {
		log.Printf("Mounting local directory %s", p.GrafanaDir)
		src, err = containers.MountLocalDir(client, p.GrafanaDir)
		if err != nil {
			return nil, err
		}
	}

	// If the enterprise global flag is set, then clone and initialize Grafana Enterprise as well.
	if p.BuildEnterprise {
		log.Println("We will build the enterprise version of Grafana")

		if p.EnterpriseDir != "" {
			log.Printf("Mounting local enterprise directory %s", p.EnterpriseDir)
			enterpriseSrcDir, err := containers.MountLocalDir(client, p.EnterpriseDir)
			if err != nil {
				return nil, err
			}
			src = containers.InitializeEnterprise(client, src, enterpriseSrcDir)
		} else {
			log.Printf("Cloning Grafana Enterprise repo from https://github.com/grafana/grafana-enterprise.git, ref %s", p.EnterpriseRef)
			enterpriseSrcDir, err := containers.CloneWithGitHubToken(client, p.GitHubToken, "https://github.com/grafana/grafana-enterprise.git", p.EnterpriseRef)
			if err != nil {
				return nil, err
			}
			src = containers.InitializeEnterprise(client, src, enterpriseSrcDir)
		}
	}
	// return src, fmt.Errorf("this is the end my friend")
	return src, nil
}
