package main

import (
	"github.com/grafana/grafana-build/artifacts"
	"github.com/urfave/cli/v2"
)

type CLI struct {
	artifacts map[string]artifacts.Initializer
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

			// Legacy commands, should eventually be completely replaced by what's in "artifacts"
			{
				Name: "package",
				Subcommands: []*cli.Command{
					PackagePublishCommand,
				},
			},
			{
				Name: "docker",
				Subcommands: []*cli.Command{
					DockerPublishCommand,
				},
			},
			ProImageCommand,
			{
				Name: "npm",
				Subcommands: []*cli.Command{
					PublishNPMCommand,
				},
			},
			GCOMCommand,
		},
	}
}

func (c *CLI) Register(flag string, a artifacts.Initializer) error {
	c.artifacts[flag] = a
	return nil
}

func (c *CLI) Initializers() map[string]artifacts.Initializer {
	return c.artifacts
}

var globalCLI = &CLI{
	artifacts: map[string]artifacts.Initializer{},
}
