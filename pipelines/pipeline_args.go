// package pipelines has functions and types that orchestrate containers.
package pipelines

import (
	"context"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/cliutil"
	"github.com/grafana/grafana-build/containers"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

	// NPMOpts will be populated if NPMFlags are enabled on the current sub-command.
	NPMOpts *containers.NPMOpts
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
		NPMOpts:          containers.NPMOptsFromFlags(c),
	}, nil
}

// InjectPipelineArgsIntoSpan is used to copy some of the arguments passed to
// the pipeline into a top-level OpenTelemtry span. Fields that might contain
// secrets are left out.
func InjectPipelineArgsIntoSpan(span trace.Span, args PipelineArgs) {
	attributes := make([]attribute.KeyValue, 0, 10)
	attributes = append(attributes, attribute.String("platform", string(args.Platform)))
	if args.GrafanaOpts != nil {
		attributes = append(attributes, attribute.String("go-version", args.GrafanaOpts.GoVersion))
		attributes = append(attributes, attribute.String("version", args.GrafanaOpts.Version))
		attributes = append(attributes, attribute.String("grafana-dir", args.GrafanaOpts.GrafanaDir))
		attributes = append(attributes, attribute.String("grafana-ref", args.GrafanaOpts.GrafanaRef))
		attributes = append(attributes, attribute.String("enterprise-dir", args.GrafanaOpts.EnterpriseDir))
		attributes = append(attributes, attribute.String("enterprise-ref", args.GrafanaOpts.EnterpriseRef))
	}
	if args.PackageOpts != nil {
		distros := []string{}
		for _, distro := range args.PackageOpts.Distros {
			distros = append(distros, string(distro))
		}
		attributes = append(attributes, attribute.StringSlice("package-distros", distros))
	}
	span.SetAttributes(attributes...)
}
