package main

import (
	"github.com/grafana/grafana-build/cmd/artifacts"
	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
)

type CLI struct {
	artifacts []*pipeline.Artifact
	app       *cli.App
}

func (c *CLI) ArtifactsCommand() *cli.Command {
	f := artifacts.ArtifactFlags(c.artifacts)
	flags := make([]cli.Flag, len(f))
	copy(flags, f)
	return &cli.Command{
		Name:  "artifacts",
		Usage: "Use this command to declare a list of artifacts to be built and/or published",
		Flags: flags,
	}
}

func (c *CLI) App() *cli.App {
	artifactsCommand := c.ArtifactsCommand()

	return &cli.App{
		Name:  "grafana-build",
		Usage: "A build tool for Grafana",
		Commands: []*cli.Command{
			BackendCommands,
			PackageCommand,
			DebCommand,
			RPMCommand,
			CDNCommand,
			DockerCommand,
			WindowsInstallerCommand,
			ZipCommand,
			ValidateCommand,
			ProImageCommand,
			StorybookCommand,
			NPMCommand,
			artifactsCommand,
		},
	}
}

func (c *CLI) Register(a *pipeline.Artifact) error {
	c.artifacts = append(c.artifacts, a)
	return nil
}

var globalCLI = &CLI{}
