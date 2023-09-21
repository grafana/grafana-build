package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var RPMPublishCommand = &cli.Command{
	Name:   "publish",
	Action: PipelineActionWithPackageInput(pipelines.RPMPublish),
	Usage:  "Using a grafana.rpm as input (ideally one built using the 'rpm' command), publish the package to our YUM repository",
	Flags: JoinFlagsWithDefault(
		PackagePublishFlags,
		GCPFlags,
	),
}
