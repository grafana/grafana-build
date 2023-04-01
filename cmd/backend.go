package main

import (
	"github.com/grafana/grafana-build/pipelines"
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

var TestBackend = &cli.Command{
	Name:   "test",
	Action: PipelineAction(pipelines.GrafanaBackendTests),
	Flags: []cli.Flag{
		FlagUnit,
		FlagIntegration,
		FlagDatabase,
	},
}

var BuildBackend = &cli.Command{
	Name:   "build",
	Action: PipelineAction(pipelines.GrafanaBackendBuild),
	Flags: []cli.Flag{
		FlagDistros,
	},
}

var BackendCommands = []*cli.Command{TestBackend, BuildBackend}
