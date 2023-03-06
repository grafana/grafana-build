package pipelines

import (
	"context"

	"dagger.io/dagger"
	"github.com/urfave/cli/v2"
)

type PipelineArgs struct {
	// These arguments are ones that are available at the global level.
	Path    string
	Verbose bool

	// Context is available for all sub-commands that define their own flags.
	Context *cli.Context
}

type PipelineFunc func(context.Context, *dagger.Client, PipelineArgs) error
