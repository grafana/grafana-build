package artifacts

import (
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
			log = slog.New(slog.NewTextHandler(os.Stderr, nil))
		)

		client, err := dagger.Connect(ctx)
		if err != nil {
			return err
		}

		state := &pipeline.State{
			Log:        log,
			Client:     client,
			CLIContext: c,
		}

		// Build and/or publish each artifact.
		for i, v := range artifacts {
			filename, err := v.FileNameFunc(ctx, v, state)
			if err != nil {
				return fmt.Errorf("error processing artifact string '%s': %w", artifactStrings[i], err)
			}
			log.Info("Building a package with arguments", "package", filename)
		}

		return nil
	}
}
