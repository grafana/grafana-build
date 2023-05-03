package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var RPMCommand = &cli.Command{
	Name:   "rpm",
	Action: PipelineActionWithPackageInput(pipelines.RPM),
	Usage:  "Using a grafana.tar.gz as input (ideally one built using the 'package' command), create a .rpm and checksum",
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		PublishFlags,
	),
}
