package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var DebCommand = &cli.Command{
	Name:        "deb",
	Action:      PipelineActionWithPackageInput(pipelines.Deb),
	Usage:       "Using a grafana.tar.gz as input (ideally one built using the 'package' command), create a .deb and checksum",
	Subcommands: []*cli.Command{DebPublishCommand},
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		PublishFlags,
		GCPFlags,
		ConcurrencyFlags,
	),
}
