package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var BuildBackend = &cli.Command{
	Name:   "build",
	Action: PipelineAction(pipelines.GrafanaBackendBuild),
	Flags: JoinFlagsWithDefault(
		GrafanaFlags,
		PackageFlags,
	),
}

var BackendCommands = &cli.Command{
	Name:        "backend",
	Usage:       "Grafana Backend (Golang) operations",
	Subcommands: []*cli.Command{BuildBackend},
}
