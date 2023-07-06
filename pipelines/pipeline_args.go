// package pipelines has functions and types that orchestrate containers.
package pipelines

import (
	"context"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/cliutil"
	"github.com/grafana/grafana-build/containers"
)

type PipelineFunc func(context.Context, *dagger.Client, *dagger.Directory, PipelineArgs) error
type PipelineFuncWithPackageInput func(context.Context, *dagger.Client, PipelineArgs) error

type ConcurrencyOpts struct {
	Parallel int64
}

func ConcurrencyOptsFromFlags(c cliutil.CLIContext) *ConcurrencyOpts {
	return &ConcurrencyOpts{
		Parallel: c.Int64("parallel"),
	}
}

type PipelineArgs struct {
	// These arguments are ones that are available at the global level.
	Verbose bool

	// Platform, where applicable, specifies what platform (linux/arm64, for example) to run the dagger containers on.
	// This should really only be used if you know what you're doing. misusing this flag can result in really slow builds.
	// Some example scenarios where you might want to use this:
	// * You're on linux/amd64 and you're building a docker image for linux/armv7 or linux/arm64
	// * You're on linux/arm64 and you're building a package for linux/arm64
	Platform dagger.Platform

	// Context is available for all sub-commands that define their own flags.
	Context cliutil.CLIContext

	// GrafanaOpts will be populated if the GrafanaFlags are enabled on the current sub-command.
	GrafanaOpts *containers.GrafanaOpts

	// PackageOpts will be populated if the PackageFlags are enabled on the current sub-command.
	PackageOpts *containers.PackageOpts

	// PublishOpts will be populated if the PublishFlags flags are enabled on the current sub-command
	// This is set for pipelines that publish artifacts.
	PublishOpts *containers.PublishOpts

	// PackageInputOpts will be populated if the PackageInputFlags are enabled on current sub-command.
	// This is set for pipelines that accept a package as input.
	PackageInputOpts *containers.PackageInputOpts
	GPGOpts          *containers.GPGOpts
	DockerOpts       *containers.DockerOpts
	GCPOpts          *containers.GCPOpts
	ConcurrencyOpts  *ConcurrencyOpts

	// ProImageOpts will be populated if ProImageFlags are enabled on the current sub-command.
	ProImageOpts *containers.ProImageOpts
}

// PipelineArgsFromContext populates a pipelines.PipelineArgs from a CLI context.
func PipelineArgsFromContext(ctx context.Context, c cliutil.CLIContext) (PipelineArgs, error) {
	// Global flags
	var (
		verbose  = c.Bool("v")
		platform = c.String("platform")
	)
	grafanaOpts, err := containers.GrafanaOptsFromFlags(ctx, c)
	if err != nil {
		return PipelineArgs{}, err
	}

	return PipelineArgs{
		Context:          c,
		Verbose:          verbose,
		Platform:         dagger.Platform(platform),
		GrafanaOpts:      grafanaOpts,
		GPGOpts:          containers.GPGOptsFromFlags(c),
		PackageOpts:      containers.PackageOptsFromFlags(c),
		PublishOpts:      containers.PublishOptsFromFlags(c),
		PackageInputOpts: containers.PackageInputOptsFromFlags(c),
		DockerOpts:       containers.DockerOptsFromFlags(c),
		GCPOpts:          containers.GCPOptsFromFlags(c),
		ConcurrencyOpts:  ConcurrencyOptsFromFlags(c),
		ProImageOpts:     containers.ProImageOptsFromFlags(c),
	}, nil
}
