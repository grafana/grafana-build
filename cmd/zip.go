package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var ZipCommand = &cli.Command{
	Name:   "zip",
	Action: PipelineActionWithPackageInput(pipelines.Zip),
	Usage:  "Using a grafana.tar.gz as input (ideally one built using the 'package' command), create a .zip and checksum",
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		PublishFlags,
		GCPFlags,
	),
}
