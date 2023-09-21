package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var DebPublishCommand = &cli.Command{
	Name:   "publish",
	Action: PipelineActionWithPackageInput(pipelines.DebPublish),
	Usage:  "Using a grafana.deb as input (ideally one built using the 'deb' command), publish the package to our APT repository",
	Flags: JoinFlagsWithDefault(
		PackagePublishFlags,
		GCPFlags,
	),
}
