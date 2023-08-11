package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var DockerPublishCommand = &cli.Command{
	Name:   "publish",
	Action: PipelineActionWithPackageInput(pipelines.DockerPublish),
	Usage:  "Using a grafana.docker.tar.gz as input (ideally one built using the 'package' command), publish a docker image and manifest",
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		PublishFlags,
		DockerPublishFlags,
		DockerFlags,
		GCPFlags,
		ConcurrencyFlags,
	),
}
