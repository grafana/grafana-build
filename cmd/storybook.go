package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var StorybookCommand = &cli.Command{
	Name:   "storybook",
	Action: PipelineActionWithPackageInput(pipelines.Storybook),
	Usage:  "Using a grafana.tar.gz as input (ideally one built using the 'package' command), take the storybook files and upload them to the destination. This can be used to put Grafana's storybook assets into a bucket for external use.",
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		PublishFlags,
		GCPFlags,
	),
}
