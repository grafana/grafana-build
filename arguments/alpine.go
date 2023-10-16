package arguments

import (
	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
)

var AlpineImageFlag = &cli.StringFlag{
	Name:  "alpine-image",
	Usage: "This flag sets the base image of the container used to build the Grafana backend binaries",
	Value: "golang:1.20.1-alpine",
}

var AlpineImage = pipeline.Argument{
	Name:        "alpine-image",
	Description: "The grafana backend binaries ('grafana', 'grafana-cli', 'grafana-server') in a directory",
	Flags: []cli.Flag{
		GoImageFlag,
	},
	ValueFunc: nil,
}
