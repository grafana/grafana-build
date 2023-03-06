package main

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

func PipelineArgsFromContext(ctx context.Context, c *cli.Context, client *dagger.Client) (pipelines.PipelineArgs, error) {
	args := pipelines.PipelineArgs{}
	args.Verbose = c.Bool("v")
	args.Version = c.String("version")

	path := c.Args().Get(0)
	if path == "" {
		path = ".grafana"
	}

	f, err := os.Stat(path)
	// It's okay if the folder doesn't exist; if it doesn't, we'll just clone the repo.
	// Other errors though it's worth just returning on.
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return pipelines.PipelineArgs{}, err
	}

	// By default we should assume that we want to clone Grafana.
	dir, err := containers.Clone(ctx, client, "https://github.com/grafana/grafana.git", args.Version)
	if err != nil {
		return pipelines.PipelineArgs{}, err
	}

	// If it does exist but it's not a directory then we should throw an error.
	// If it doesn't exist, then this block will be skipped and the project will be cloned.
	if f != nil {
		if !f.IsDir() {
			return pipelines.PipelineArgs{}, errors.New("path provided is not a directory")
		}

		dir = client.Host().Directory(path)
	}

	args.Context = c
	args.Grafana = dir

	return args, nil
}

func PipelineAction(pf pipelines.PipelineFunc) cli.ActionFunc {
	return func(c *cli.Context) error {
		var (
			ctx  = c.Context
			opts = []dagger.ClientOpt{}
		)
		if c.Bool("verbose") {
			opts = append(opts, dagger.WithLogOutput(os.Stderr))
		}
		client, err := dagger.Connect(ctx, opts...)
		if err != nil {
			return err
		}

		args, err := PipelineArgsFromContext(ctx, c, client)
		if err != nil {
			return err
		}

		return pf(c.Context, client, args)
	}
}

var app = &cli.App{
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Value:   false,
		},
		&cli.StringFlag{
			Name:     "version",
			Required: false,
			Value:    "main",
		},
	},
	Commands: []*cli.Command{
		{
			Name:        "backend",
			Usage:       "Grafana Backend (Golang) operations",
			Subcommands: BackendCommands,
		},
	},
}

func main() {

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
