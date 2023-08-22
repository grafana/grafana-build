package main

import (
	"github.com/grafana/grafana-build/cmd/artifacts"
	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
)

func artifactsAction(a ...*pipeline.Artifact) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		return nil
	}
}

var ArtifactsCommand = &cli.Command{
	Name: "artifacts",
	Action: artifactsAction(
		artifacts.Backend,
		artifacts.Frontend,
		artifacts.Tarball,
	),
	Usage: "Temporary: Select from a list of artifacts to build or publish",
}
