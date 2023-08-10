package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var DockerCommand = &cli.Command{
	Name:        "docker",
	Action:      PipelineActionWithPackageInput(pipelines.Docker),
	Usage:       "Using a grafana.tar.gz as input (ideally one built using the 'package' command), create a docker image",
	Subcommands: []*cli.Command{DockerPublishCommand},
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		PublishFlags,
		DockerFlags,
		GCPFlags,
		ConcurrencyFlags,
	),
}
