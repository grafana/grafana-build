package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var FlagUnit = &cli.BoolFlag{
	Name:  "unit",
	Usage: "Run the backend unit tests",
	Value: true,
}

var FlagIntegration = &cli.BoolFlag{
	Name:  "integration",
	Usage: "Run the backend integration tests",
	Value: false,
}

var FlagDatabase = &ChoiceFlag{
	Name:    "database",
	Usage:   "Which database to use, only valid when --integration=true",
	Choices: pipelines.IntegrationDatabases,
	Value:   "sqlite",
}

var BackendTestFlags = []cli.Flag{
	FlagUnit,
	FlagIntegration,
	FlagDatabase,
}

var TestBackend = &cli.Command{
	Name:   "test",
	Action: PipelineAction(pipelines.GrafanaBackendTests),
	Flags: JoinFlagsWithDefault(
		GrafanaFlags,
		BackendTestFlags,
	),
}

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
	Subcommands: []*cli.Command{TestBackend, BuildBackend},
}
