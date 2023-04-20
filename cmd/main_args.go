package main

import (
	"context"

	"github.com/grafana/grafana-build/pipelines"
)

// PipelineArgsFromContext populates a pipelines.PipelineArgs from a CLI context.
func PipelineArgsFromContext(ctx context.Context, c pipelines.CLIContext) (pipelines.PipelineArgs, error) {
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
		enterprise = true
	}

	//if version == "" {
	//	log.Println("Version not provided; getting version from package.json...")
	//	v, err := containers.GetPackageJSONVersion(ctx, client, src)
	//	if err != nil {
	//		return pipelines.PipelineArgs{}, err
	//	}

	//	version = v
	//	log.Println("Got version", v)
	//}

	return pipelines.PipelineArgs{
		BuildID:         buildID,
		Verbose:         verbose,
		ProvidedVersion: version,
		BuildEnterprise: enterprise,
		BuildGrafana:    grafana,
		GrafanaDir:      grafanaDir,
		GrafanaRef:      ref,
		EnterpriseDir:   enterpriseDir,
		EnterpriseRef:   enterpriseRef,
		Context:         c,
		GitHubToken:     gitHubToken,
		//Grafana:         src,
	}, nil
}
