package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

func PipelineArgsFromContext(c *cli.Context, client *dagger.Client) (pipelines.PipelineArgs, error) {
	var (
		verbose    = c.Bool("v")
		version    = c.String("version")
		enterprise = c.Bool("enterprise")
	)

	path := c.Args().Get(0)
	if path == "" {
		path = ".grafana"
	}

	f, err := os.Stat(path)
	// It's okay if the folder doesn't exist; if it doesn't, we'll just clone the repo.
	// Other errors though it's worth just returning on.
	if err == nil {
		// If it does exist but it's not a directory then we should throw an error.
		// If it doesn't exist, then this block will be skipped and the project will be cloned.
		if !f.IsDir() {
			return pipelines.PipelineArgs{}, errors.New("path provided is not a directory")
		}

		return pipelines.PipelineArgs{
			Verbose:    verbose,
			Version:    version,
			Enterprise: enterprise,
			Context:    c,
			Grafana:    client.Host().Directory(path),
		}, nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		return pipelines.PipelineArgs{}, err
	}

	// If the folder doesn't exist, then we want to clone Grafana.
	src, err := containers.Clone(client, "https://github.com/grafana/grafana.git", version)
	if err != nil {
		return pipelines.PipelineArgs{}, err
	}

	// If the 'enterprise global flag is set, then clone and initialize Grafana Enterprise as well.
	if enterprise {
		enterpriseDir, err := containers.CloneWithSSHAuth(client, filepath.Clean(os.Getenv("HOME")+"/.ssh/id_rsa"), "git@github.com:grafana/grafana-enterprise.git", version)
		if err != nil {
			return pipelines.PipelineArgs{}, err
		}

		src = containers.InitializeEnterprise(client, src, enterpriseDir)
	}

	return pipelines.PipelineArgs{
		Verbose:    verbose,
		Version:    version,
		Enterprise: enterprise,
		Context:    c,
		Grafana:    src,
	}, nil
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

		args, err := PipelineArgsFromContext(c, client)
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
		&cli.BoolFlag{
			Name:  "enterprise",
			Usage: "If set, attempt to clone and initialize Grafana Enterprise",
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
		PackageCommand,
	},
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
