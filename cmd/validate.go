package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var ValidateCommand = &cli.Command{
	Name:        "validate",
	Action:      PipelineAction(pipelines.ValidatePackage),
	Description: "Validates a grafana.tar.gz for the given distributions (--distro) placed in the destination directory (--destination)",
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		GrafanaFlags,
		GCPFlags,
		ConcurrencyFlags,
	),
}
