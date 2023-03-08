package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var PackageCommand = &cli.Command{
	Name:   "package",
	Usage:  "Creates a grafana.tar.gz in the current working directory",
	Action: PipelineAction(pipelines.Package),
	Flags: []cli.Flag{
		FlagDistro,
	},
}
