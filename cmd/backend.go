package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var TestBackendUnit = &cli.Command{
	Name:   "test",
	Action: PipelineAction(pipelines.GrafanaBackendTests),
}

var TestBackendIntegration = &cli.Command{
	Name:   "test-integration",
	Action: PipelineAction(pipelines.GrafanaBackendTestIntegration),
}
