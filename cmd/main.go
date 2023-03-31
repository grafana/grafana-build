package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"time"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

func init() {
	rand.Seed(time.Now().Unix())
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
			Name:     "grafana-ref",
			Required: false,
			Value:    "main",
		},
		&cli.StringFlag{
			Name:     "enterprise-ref",
			Required: false,
			Value:    "main",
		},
		&cli.StringFlag{
			Name:     "github-token",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "build-id",
			Required: false,
		},
	},
	Before: func(cctx *cli.Context) error {
		token, err := lookupGitHubToken(cctx)
		if err != nil {
			return fmt.Errorf("failed to find a GitHub access token: %w", err)
		}
		if token == "" {
			return fmt.Errorf("could not find a GitHub token in the environment")
		}
		return cctx.Set("github-token", token)
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

func PipelineArgsFromContext(c *cli.Context, client *dagger.Client) (pipelines.PipelineArgs, error) {
	var (
		verbose       = c.Bool("v")
		version       = c.String("version")
		ref           = c.String("grafana-ref")
		enterprise    = c.Bool("enterprise")
		enterpriseRef = c.String("enterprise-ref")
		buildID       = c.String("build-id")
	)

	if buildID == "" {
		buildID = randomString(12)
	}

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
			BuildID:       buildID,
			Verbose:       verbose,
			Version:       version,
			Ref:           ref,
			EnterpriseRef: enterpriseRef,
			Enterprise:    enterprise,
			Context:       c,
			Grafana:       client.Host().Directory(path),
		}, nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		return pipelines.PipelineArgs{}, err
	}

	// If the folder doesn't exist, then we want to clone Grafana.
	src, err := containers.Clone(client, "https://github.com/grafana/grafana.git", ref)
	if err != nil {
		return pipelines.PipelineArgs{}, err
	}

	// If the 'enterprise global flag is set, then clone and initialize Grafana Enterprise as well.
	if enterprise {
		enterpriseDir, err := containers.CloneWithGitHubToken(client, c.String("github-token"), "https://github.com/grafana/grafana-enterprise.git", enterpriseRef)
		if err != nil {
			return pipelines.PipelineArgs{}, err
		}

		src = containers.InitializeEnterprise(client, src, enterpriseDir)
	}

	return pipelines.PipelineArgs{
		BuildID:    buildID,
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

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
