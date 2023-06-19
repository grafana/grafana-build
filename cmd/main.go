package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/pipelines"
	"github.com/urfave/cli/v2"
)

var app = &cli.App{
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
	},
	Action: MainCommand,
	Flags: JoinFlagsWithDefault(GrafanaFlags, PackageFlags, PublishFlags, DockerFlags, []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "artifact",
			Usage: "Specify the output artifact of the command (deb|rpm|docker|tarball)",
		},
	}),
}

func buildArtifact(ctx context.Context, cache map[string]*dagger.Directory, cacheKey string, art string, reg *pipelines.ArtifactDefinitionRegistry, d *dagger.Client, src *dagger.Directory, genOpts pipelines.ArtifactGeneratorOptions) (*dagger.Directory, error) {
	cached, inCache := cache[cacheKey]
	if inCache {
		return cached, nil
	}
	def, ok := reg.Get(art)
	if !ok {
		return nil, fmt.Errorf("could not resolve artifact `%s`", art)
	}
	mounts := make(map[string]*dagger.Directory)
	for k, v := range def.Requirements {
		_, ok := reg.Get(v)
		if !ok {
			return nil, fmt.Errorf("could not resolve dependency of `%s`: %s", art, v)
		}

		subOut, err := buildArtifact(ctx, cache, cacheKey, v, reg, d, src, genOpts)
		if err != nil {
			return nil, fmt.Errorf("could not build `%s->%s`: %w", art, v, err)
		}
		mounts[k] = subOut
	}

	result, err := def.Generator(ctx, d, src, genOpts, mounts)
	if err != nil {
		return nil, err
	}
	cache[cacheKey] = result
	return result, nil
}

func MainCommand(cliCtx *cli.Context) error {
	artifacts := cliCtx.StringSlice("artifact")
	if len(artifacts) == 0 {
		return fmt.Errorf("specify at least one artifact")
	}
	for _, artifact := range artifacts {
		_, ok := pipelines.DefaultArtifacts.Get(artifact)
		if !ok {
			return fmt.Errorf("unsupported artifact requested: %s", artifact)
		}
	}
	return PipelineAction(func(ctx context.Context, d *dagger.Client, src *dagger.Directory, args pipelines.PipelineArgs) error {
		results := make(map[string]*dagger.Directory)
		cache := make(map[string]*dagger.Directory)

		finalArtifacts, err := pipelines.GenerateFinalArtifactList(ctx, d, pipelines.DefaultArtifacts, artifacts, args)
		if err != nil {
			return err
		}

		log.Printf("Generating the following artifacts: %v", finalArtifacts)

		if len(finalArtifacts) == 0 {
			return fmt.Errorf("the combination of distros and artifacts didn't produce any valid output artifacts")
		}

		for _, req := range finalArtifacts {
			artifact := req.Name
			distro := req.Options.Distribution
			cacheKey := fmt.Sprintf("%s-%s", artifact, distro)
			dir, err := buildArtifact(ctx, cache, cacheKey, artifact, pipelines.DefaultArtifacts, d, src, req.Options)
			if err != nil {
				return err
			}
			results[cacheKey] = dir
		}

		dest := cliCtx.String("destination")
		for _, req := range finalArtifacts {
			distro := req.Options.Distribution
			artifact := req.Name
			cacheKey := fmt.Sprintf("%s-%s", artifact, distro)
			dir := results[cacheKey]
			if _, err := containers.PublishDirectory(ctx, d, dir, args.GCPOpts, dest); err != nil {
				return err
			}
		}
		return nil
	})(cliCtx)
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

		args, err := pipelines.PipelineArgsFromContext(c.Context, c)
		if err != nil {
			return err
		}

		grafanaDir, err := args.GrafanaOpts.Grafana(ctx, client)
		if err != nil {
			return err
		}

		v, err := args.GrafanaOpts.DetectVersion(ctx, client, grafanaDir)
		if err != nil {
			return err
		}
		args.GrafanaOpts.Version = v

		return pf(c.Context, client, grafanaDir, args)
	}
}

func PipelineActionWithPackageInput(pf pipelines.PipelineFuncWithPackageInput) cli.ActionFunc {
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

		args, err := pipelines.PipelineArgsFromContext(c.Context, c)
		if err != nil {
			return err
		}

		if len(args.PackageInputOpts.Packages) == 0 {
			return errors.New("expected at least one package from a '--package' flag")
		}

		return pf(c.Context, client, args)
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
