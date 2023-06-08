package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var CDNCommand = &cli.Command{
	Name:   "cdn",
	Action: PipelineActionWithPackageInput(pipelines.CDN),
	Usage:  "Using a grafana.tar.gz as input (ideally one built using the 'package' command), take the frontend files and upload them to the destination. This can be used to put Grafana's frontend assets into a bucket for use in a CDN.",
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		PublishFlags,
		GCPFlags,
	),
}
