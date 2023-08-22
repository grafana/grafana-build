package artifacts

import (
	"context"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/cliutil"
	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
)

var GoImageFlag = &cli.StringFlag{
	Name:     "go-image",
	Usage:    "This flag sets the base image of the container used to build the Grafana backend binaries",
	Required: true,
	Value:    "golang:1.20.1-alpine",
}

var ArgumentGoImage = pipeline.Argument{
	Name:        "go-image",
	Description: "The grafana backend binaries ('grafana', 'grafana-cli', 'grafana-server') in a directory",
	Flags: []cli.Flag{
		GoImageFlag,
	},
	ValueFunc: func(ctx context.Context, c cliutil.CLIContext, d *dagger.Client) (any, error) {
		return c.String("go-image"), nil
	},
	Required: false,
}
