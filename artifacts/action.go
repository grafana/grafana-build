package artifacts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
)

// Register adds the pipeline artifacts to the urfave/cli app.
func Register(r Registerer, a ...pipeline.Artifact) error {
	for _, v := range a {
		if err := r.Register(v); err != nil {
			return err
		}
	}

	return nil
}

func Command(r Registerer) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		// ArtifactStrings represent an artifact with a list of boolean options, like
		// targz:linux/amd64:enterprise
		artifactStrings := c.StringSlice("artifacts")

		if len(artifactStrings) == 0 {
			return errors.New("no artifacts specified. At least 1 artifact is required using the '--artifact' or '-a' flag")
		}

		registered := r.Artifacts()
		// Get the artifacts that were specified by the artifacts commands.
		// These are specified by using artifact strings, or comma-delimited lists of flags.
		artifacts, err := ArtifactsFromStrings(artifactStrings, registered)
		if err != nil {
			return err
		}

		var (
			ctx = c.Context
			log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))
		)

		log.Debug("Connecting to dagger daemon...")
		client, err := dagger.Connect(ctx)
		if err != nil {
			return err
		}
		log.Debug("Connected to dagger daemon")

		state := &pipeline.State{
			Log:        log,
			Client:     client,
			CLIContext: c,
		}

		opts := &pipeline.ArtifactContainerOpts{
			Client:   client,
			Log:      log,
			State:    state,
			Platform: dagger.Platform(c.String("platform")),
		}
		// Build and/or publish each artifact.
		for i, v := range artifacts {
			filename, err := v.FileNameFunc(ctx, v, state)
			if err != nil {
				return fmt.Errorf("error processing artifact string '%s': %w", artifactStrings[i], err)
			}
			log.Info("Building a package with arguments", "package", filename)
			if err := BuildArtifact(ctx, v, filename, opts); err != nil {
				return err
			}
		}

		return nil
	}
}

func BuildArtifact(ctx context.Context, a pipeline.Artifact, path string, opts *pipeline.ArtifactContainerOpts) error {
	switch a.Type {
	case pipeline.ArtifactTypeDirectory:
		dir, err := BuildArtifactDirectory(ctx, a)
		if err != nil {
			return err
		}

		if _, err := dir.Export(ctx, path); err != nil {
			return err
		}
	case pipeline.ArtifactTypeFile:
		file, err := BuildArtifactFile(ctx, a, opts)
		if err != nil {
			return err
		}

		if _, err := file.Export(ctx, path); err != nil {
			return err
		}
	}

	return errors.New("unrecognized artifact type")
}

func Dependencies(a []pipeline.Artifact) map[string]pipeline.Artifact {
	d := make(map[string]pipeline.Artifact, len(a))
	for _, v := range a {
		d[v.Name] = v
	}

	return d
}

func BuildArtifactFile(ctx context.Context, a pipeline.Artifact, opts *pipeline.ArtifactContainerOpts) (*dagger.File, error) {
	log := opts.Log.With("artifact", a.Name)

	log.Debug("Getting builder...")
	builder, err := a.Builder(ctx, opts)
	if err != nil {
		return nil, err
	}
	log.Debug("Got builder")

	return a.BuildFileFunc(ctx, &pipeline.ArtifactBuildOpts{
		ContainerOpts: opts,
		Builder:       builder,
		Dependencies:  Dependencies(a.Requires),
	})
}

func BuildArtifactDirectory(ctx context.Context, a pipeline.Artifact) (*dagger.Directory, error) {
	return nil, nil
}
