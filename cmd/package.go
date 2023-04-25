package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var PackageCommand = &cli.Command{
	Name: "package",
	Subcommands: []*cli.Command{
		{
			Name:   "build",
			Usage:  "Creates a grafana.tar.gz in the current working directory",
			Action: PipelineAction(pipelines.Package),
			Flags: JoinFlagsWithDefault(
				GrafanaFlags,
				PackageFlags,
			),
		},
		{
			Name:   "publish",
			Action: PipelineAction(pipelines.PublishPackage),
			Flags: JoinFlagsWithDefault(
				GrafanaFlags,
				PackageFlags,
				PublishFlags,
			),
		},
	},
}
