package pipelines

import (
	"context"

	"dagger.io/dagger"
)

type PipelineArgs struct {
	Path    string
	Verbose bool
}

type PipelineFunc func(context.Context, *dagger.Client, PipelineArgs) error
