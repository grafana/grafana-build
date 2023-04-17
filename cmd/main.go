package main

import (
	"fmt"
	"log"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:     "grafana",
			Usage:    "If set, initialize Grafana",
			Required: false,
			Value:    true,
		},
		&cli.StringFlag{
			Name:     "grafana-dir",
			Usage:    "Local Grafana dir to use, instead of git clone",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "grafana-ref",
			Usage:    "Grafana ref to clone, not valid if --grafana-dir is set",
			Required: false,
			Value:    "main",
		},
		&cli.BoolFlag{
			Name:  "enterprise",
			Usage: "If set, initialize Grafana Enterprise",
			Value: false,
		},
		&cli.StringFlag{
			Name:     "enterprise-dir",
			Usage:    "Local Grafana Enterprise dir to use, instead of git clone",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "enterprise-ref",
			Usage:    "Grafana Enterprise ref to clone, not valid if --enterprise-dir is set",
			Required: false,
			Value:    "main",
		},
		&cli.StringFlag{
			Name:     "build-id",
			Usage:    "Build ID to use, by default will be what is defined in package.json",
			Required: false,
		},
		&cli.StringFlag{
			Name:     "github-token",
			Usage:    "Github token to use for git cloning, by default will be pulled from GitHub",
			Required: false,
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "Increase log verbosity",
			Value:   false,
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

func PipelineArgsFromContext(c *cli.Context, client *dagger.Client) (pipelines.PipelineArgs, error) {
	var (
		verbose       = c.Bool("v")
		version       = c.String("version")
		grafana       = c.Bool("grafana")
		grafanaDir    = c.String("grafana-dir")
		ref           = c.String("grafana-ref")
		enterprise    = c.Bool("enterprise")
		enterpriseDir = c.String("enterprise-dir")
		enterpriseRef = c.String("enterprise-ref")
		buildID       = c.String("build-id")
		src           *dagger.Directory
		err           error
	)

	if buildID == "" {
		buildID = randomString(12)
	}

	if enterpriseRef != "main" {
		enterprise = true
	}

	if enterpriseDir != "" {
		_, err := os.Stat(enterpriseDir)
		if err != nil {
			return pipelines.PipelineArgs{}, fmt.Errorf("--enterprise-dir %s does not exist", enterpriseDir)
		}
		enterprise = true
	}

	// Determine if we need to git authenticate, and then do it
	do_git_auth := false
	if grafanaDir == "" {
		// Grafana will clone from git, needs auth
		do_git_auth = true
	} else if enterprise == true && enterpriseDir == "" {
		// Enterprise will clone from git, needs auth
		do_git_auth = true
	}
	if do_git_auth {
		log.Println("Aquiring github token")
		token, err := lookupGitHubToken(c)
		if err != nil {
			return pipelines.PipelineArgs{}, err
		}
		if token == "" {
			return pipelines.PipelineArgs{}, fmt.Errorf("Unable to aquire github token")
		}
		c.Set("github-token", token)
	}

	// Start by gathering grafana code, either locally or from git
	if grafanaDir != "" {
		log.Printf("Mounting local directory %s", grafanaDir)
		src, err = containers.MountLocalDir(client, grafanaDir)
		if err != nil {
			return pipelines.PipelineArgs{}, err
		}
	} else {
		log.Printf("Cloning Grafana repo from https://github.com/grafana/grafana.git, ref %s", ref)
		src, err = containers.Clone(client, "https://github.com/grafana/grafana.git", ref)
		if err != nil {
			return pipelines.PipelineArgs{}, err
		}
	}

	// If the enterprise global flag is set, then clone and initialize Grafana Enterprise as well.
	if enterprise {
		log.Println("We will build the enterprise version of Grafana")

		if enterpriseDir != "" {
			log.Printf("Mounting local enterprise directory %s", enterpriseDir)
			enterpriseSrcDir, err := containers.MountLocalDir(client, enterpriseDir)
			if err != nil {
				return pipelines.PipelineArgs{}, err
			}
			src = containers.InitializeEnterprise(client, src, enterpriseSrcDir)
		} else {
			log.Printf("Cloning Grafana Enterprise repo from https://github.com/grafana/grafana-enterprise.git, ref %s", enterpriseRef)
			enterpriseSrcDir, err := containers.CloneWithGitHubToken(client, c.String("github-token"), "https://github.com/grafana/grafana-enterprise.git", enterpriseRef)
			if err != nil {
				return pipelines.PipelineArgs{}, err
			}
			src = containers.InitializeEnterprise(client, src, enterpriseSrcDir)
		}
	}
	//return pipelines.PipelineArgs{}, fmt.Errorf("this is the end my friend")

	if version == "" {
		log.Println("Version not provided; getting version from package.json...")
		v, err := containers.GetPackageJSONVersion(c.Context, client, src)
		if err != nil {
			return pipelines.PipelineArgs{}, err
		}

		version = v
		log.Println("Got version", v)
	}

	return pipelines.PipelineArgs{
		BuildID:         buildID,
		Verbose:         verbose,
		Version:         version,
		BuildEnterprise: enterprise,
		BuildGrafana:    grafana,
		Context:         c,
		Grafana:         src,
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
