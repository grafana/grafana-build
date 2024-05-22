package arguments

import (
	"context"

	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
)

var ProDirectoryFlags = []cli.Flag{
	&cli.StringFlag{},
}

// GrafanaDirectory will provide the valueFunc that initializes and returns a *dagger.Directory that has Grafana in it.
// Where possible, when cloning and no authentication options are provided, the valuefunc will try to use the configured github CLI for cloning.
var ProDirectory = pipeline.Argument{
	Name:        "pro-dir",
	Description: "The source tree of that has the Dockerfile for Grafana Pro",
	Flags:       ProDirectoryFlags,
	ValueFunc:   proDirectory,
}

func proDirectory(ctx context.Context, opts *pipeline.ArgumentOpts) (any, error) {
	return nil, nil
}
