package main

import (
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var PackageCommand = &cli.Command{
	Name:   "package",
	Usage:  "Creates a grafana.tar.gz in the current working directory",
	Action: PipelineAction(pipelines.Package),
	Flags: []cli.Flag{
		FlagDistros,
	},
	Subcommands: []*cli.Command{
		{
			Name:   "publish",
			Action: PipelineAction(pipelines.PublishPackage),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "destination",
					Usage:    "GCS URL to upload the package (example: gs://bucket/grafana.tar.gz)",
					Aliases:  []string{"d"},
					Required: true,
				},
				&cli.StringFlag{
					Name:     "key",
					Usage:    "Provides a service-account keyfile to use to authenticate with the Google Cloud SDK. If not provided or is empty, then $XDG_CONFIG_HOME/gcloud will be mounted in the container",
					Required: false,
				},
			},
		},
	},
}
