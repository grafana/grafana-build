package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var ValidateUpgradeCommand = &cli.Command{
	Name:   "upgrade",
	Action: PipelineAction(pipelines.ValidatePackageUpgrade),
	Usage:  "Validates if a .deb or a .rpm package (--from) can be upgraded by another .deb or .rpm package (--to)",
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		GCPFlags,
	),
}

var ValidateCommand = &cli.Command{
	Name:        "validate",
	Action:      PipelineAction(pipelines.ValidatePackage),
	Description: "Validates grafana .tar.gz, .deb, .rpm and .docker.tar.gz packages and places the results in the destination directory (--destination)",
	Subcommands: []*cli.Command{ValidateUpgradeCommand},
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		GrafanaFlags,
		PublishFlags,
		GCPFlags,
		ConcurrencyFlags,
	),
}
