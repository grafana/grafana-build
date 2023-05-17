package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var WindowsInstallerCommand = &cli.Command{
	Name:   "windows-installer",
	Action: PipelineActionWithPackageInput(pipelines.WindowsInstaller),
	Usage:  "Using a grafana.tar.gz as input (ideally one built using the 'package' command), create a .exe installer and checksum",
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		PublishFlags,
		GPGFlags,
	),
}
