package arguments

import (
	"context"

	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
)

var GCPFlags = []cli.Flag{}

var GCP = pipeline.Argument{
	Name:        "gcp",
	Description: "",
	Flags:       GCPFlags,
	ValueFunc:   proDirectory,
}

func gcp(ctx context.Context, opts *pipeline.ArgumentOpts) (any, error) {
	return nil, nil
}
