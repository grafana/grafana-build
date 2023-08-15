package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var NPMCommand = &cli.Command{
	Name:   "npm",
	Action: PipelineActionWithPackageInput(pipelines.NPM),
	Usage:  "Using a grafana.tar.gz as input (ideally one built using the 'package' command), take the npm artifacts and upload them to the destination. This can be used to put Grafana's npm artifacts into a bucket for external use.",
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		PublishFlags,
		GCPFlags,
	),
}
