package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var GCOMPublishCommand = &cli.Command{
	Name:        "publish",
	Action:      PipelineActionWithPackageInput(pipelines.PublishGCOM),
	Description: "Publishes a grafana.tar.gz (ideally one built using the 'package' command) to grafana.com (--destination will be the download path)",
	Flags: JoinFlagsWithDefault(
		GCOMFlags,
		PackageInputFlags,
		PublishFlags,
		ConcurrencyFlags,
	),
}
