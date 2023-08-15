package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var PackageCommand = &cli.Command{
	Name:        "package",
	Action:      PipelineAction(pipelines.BuildPackage),
	Description: "Creates a grafana.tar.gz for the given distributions (--distro) placed in the destination directory (--destination)",
	Subcommands: []*cli.Command{PackagePublishCommand},
	Flags: JoinFlagsWithDefault(
		GrafanaFlags,
		PackageFlags,
		PublishFlags,
		GCPFlags,
		ConcurrencyFlags,
	),
}
