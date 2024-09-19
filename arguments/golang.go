package arguments

import (
	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
)

const (
	DefaultGoVersion      = "1.23.1"
	DefaultViceroyVersion = "v0.4.0"
)

var GoVersionFlag = &cli.StringFlag{
	Name:  "go-version",
	Usage: "The Go version to use when compiling Grafana",
	Value: DefaultGoVersion,
}

var GoVersion = pipeline.NewStringFlagArgument(GoVersionFlag)

var ViceroyVersionFlag = &cli.StringFlag{
	Name:  "viceroy-version",
	Usage: "This flag sets the base image of the container used to build the Grafana backend binaries for non-Linux distributions",
	Value: DefaultViceroyVersion,
}

var ViceroyVersion = pipeline.NewStringFlagArgument(ViceroyVersionFlag)
