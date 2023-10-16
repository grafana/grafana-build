package artifacts

import (
	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
)

type CLIRegisterer struct {
	command *cli.Command
}

func (c *CLIRegisterer) Register(a *pipeline.Artifact) error {
	return nil
}
