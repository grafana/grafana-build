// package pipelines has functions and types that orchestrate containers.
package pipelines

import (
	"context"

	"dagger.io/dagger"
	"github.com/urfave/cli/v2"
)

type PipelineArgs struct {
	// These arguments are ones that are available at the global level.
	Grafana         *dagger.Directory
	Verbose         bool
	Ref             string
	BuildGrafana    bool
	BuildEnterprise bool
	EnterpriseRef   string
	Version         string
	BuildID         string

	// Context is available for all sub-commands that define their own flags.
	Context *cli.Context
}

type PipelineFunc func(context.Context, *dagger.Client, PipelineArgs) error
