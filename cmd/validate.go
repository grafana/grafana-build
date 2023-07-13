package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var ValidateUpgradeCommand = &cli.Command{
	Name:   "upgrade",
	Action: PipelineAction(pipelines.ValidatePackageUpgrade),
	Usage:  "Validates if a list of .deb or .rpm packages can be upgraded to each other sequentially",
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		GCPFlags,
	),
}

var ValidateSignatureCommand = &cli.Command{
	Name:   "checksig",
	Action: PipelineAction(pipelines.ValidatePackageSignature),
	Usage:  "Validates if a .rpm package is properly signed using the provided GPG public key",
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		GPGPublicFlags,
		GCPFlags,
	),
}

var ValidateCommand = &cli.Command{
	Name:        "validate",
	Action:      PipelineAction(pipelines.ValidatePackage),
	Description: "Validates grafana .tar.gz, .deb, .rpm and .docker.tar.gz packages and places the results in the destination directory (--destination)",
	Subcommands: []*cli.Command{ValidateUpgradeCommand, ValidateSignatureCommand},
	Flags: JoinFlagsWithDefault(
		PackageInputFlags,
		GrafanaFlags,
		PublishFlags,
		GCPFlags,
		ConcurrencyFlags,
	),
}
