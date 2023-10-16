package main

import (
	"github.com/grafana/grafana-build/artifacts"
	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
)

type CLI struct {
	artifacts []pipeline.Artifact
	app       *cli.App
}

func (c *CLI) ArtifactsCommand() *cli.Command {
	f := artifacts.ArtifactFlags(c)
	flags := make([]cli.Flag, len(f))
	copy(flags, f)
	return &cli.Command{
		Name:   "artifacts",
		Usage:  "Use this command to declare a list of artifacts to be built and/or published",
		Flags:  flags,
		Action: artifacts.Command(c),
	}
}

func (c *CLI) App() *cli.App {
	artifactsCommand := c.ArtifactsCommand()

	return &cli.App{
		Name:  "grafana-build",
		Usage: "A build tool for Grafana",
		Commands: []*cli.Command{
			artifactsCommand,
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
		},
	}
}

func (c *CLI) Register(a pipeline.Artifact) error {
	c.artifacts = append(c.artifacts, a)
	return nil
}

func (c *CLI) Artifacts() []pipeline.Artifact {
	return c.artifacts
}

var globalCLI = &CLI{}
