package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var PackagePublishCommand = &cli.Command{
	Name:        "publish",
	Action:      PipelineActionWithPackageInput(pipelines.PublishPackage),
	Description: "Publishes a grafana.tar.gz (ideally one built using the 'package' command) in the destination directory (--destination)",
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		PublishFlags,
		GCPFlags,
		ConcurrencyFlags,
	),
}
