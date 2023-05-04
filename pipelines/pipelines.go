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

type PipelineArgs struct {
	// These arguments are ones that are available at the global level.
	Verbose bool

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

	// GPGOpts will be populated if the GPGFlags are enabled on the current sub-command.
	GPGOpts *containers.GPGOpts
}

// PipelineArgsFromContext populates a pipelines.PipelineArgs from a CLI context.
func PipelineArgsFromContext(ctx context.Context, c cliutil.CLIContext) (PipelineArgs, error) {
	// Global flags...
	verbose := c.Bool("v")
	grafanaOpts, err := containers.GrafanaOptsFromFlags(ctx, c)
	if err != nil {
		return PipelineArgs{}, err
	}

	return PipelineArgs{
		Context:          c,
		Verbose:          verbose,
		GrafanaOpts:      grafanaOpts,
		GPGOpts:          containers.GPGOptsFromFlags(c),
		PackageOpts:      containers.PackageOptsFromFlags(c),
		PublishOpts:      containers.PublishOptsFromFlags(c),
		PackageInputOpts: containers.PackageInputOptsFromFlags(c),
	}, nil
}
