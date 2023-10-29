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
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

func Action(r Registerer, c *cli.Context) error {
	// ArtifactStrings represent an artifact with a list of boolean options, like
	// targz:linux/amd64:enterprise
	artifactStrings := c.StringSlice("artifacts")

	var (
		ctx = c.Context
		log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		parallel    = c.Int64("parallel")
		destination = c.String("destination")
	)

	if len(artifactStrings) == 0 {
		return errors.New("no artifacts specified. At least 1 artifact is required using the '--artifact' or '-a' flag")
	}

	log.Debug("Connecting to dagger daemon...")
	client, err := dagger.Connect(ctx)
	if err != nil {
		return err
	}
	log.Debug("Connected to dagger daemon")

	var state pipeline.StateHandler = &pipeline.State{
		Log:        log,
		Client:     client,
		CLIContext: c,
	}

	registered := r.Initializers()

	log.Debug("Generating artifacts from artifact strings...")
	// Initialize the artifacts that were specified by the artifacts commands.
	// These are specified by using artifact strings, or comma-delimited lists of flags.
	artifacts, err := ArtifactsFromStrings(ctx, log, artifactStrings, registered, state)
	if err != nil {
		return err
	}
	log.Debug("Done generating artifact metadata")

	state = pipeline.StateWithLogger(
		log.With("service", "state"),
		state,
	)

	// The artifact store is responsible for storing built artifacts and issuing them to artifacts that use them as dependencies using the artifact's filename as the key.
	store := pipeline.NewArtifactStore(log)

	opts := &pipeline.ArtifactContainerOpts{
		Client:   client,
		Log:      log,
		State:    state,
		Platform: dagger.Platform(c.String("platform")),
		Store:    store,
	}

	// Build each artifact, essentially constructing a dag.
	for i, v := range artifacts {
		filename, err := v.Handler.Filename(ctx)
		if err != nil {
			return fmt.Errorf("error processing artifact string '%s': %w", artifactStrings[i], err)
		}
		log := log.With("filename", filename, "artifact", v.ArtifactString)
		log.Info("Adding artifact to dag...")
		if err := BuildArtifact(ctx, log, v, opts); err != nil {
			return err
		}
		log.Info("Done adding artifact")

	}

	checksum := c.Bool("checksum")
	wg := &errgroup.Group{}
	sm := semaphore.NewWeighted(parallel)
	log.Info("Exporting artifacts...")
	// Export the files from the dag, causing the containers to trigger.
	for _, v := range artifacts {
		log := log.With("artifact", v.ArtifactString)
		wg.Go(ExportArtifactFunc(ctx, client, sm, log, v, store, destination, checksum))
	}

	return wg.Wait()
}
func BuildArtifact(ctx context.Context, log *slog.Logger, a *pipeline.Artifact, opts *pipeline.ArtifactContainerOpts) error {
	store := opts.Store
	exists, err := store.Exists(ctx, a)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// populate the dependency list
	dependencies, err := a.Handler.Dependencies(ctx)
	if err != nil {
		return err
	}

	// Get the files / directories that the dependencies define,
	// and store the result for re-use.
	for _, v := range dependencies {
		f, err := v.Handler.Filename(ctx)
		if err != nil {
			return err
		}
		log := log.With("artifact", v.ArtifactString, "filename", f)
		if err := BuildArtifact(ctx, log, v, opts); err != nil {
			return err
		}
	}

	switch a.Type {
	case pipeline.ArtifactTypeDirectory:
		dir, err := BuildArtifactDirectory(ctx, a, opts)
		if err != nil {
			return err
		}

		return store.StoreDirectory(ctx, a, dir)
	case pipeline.ArtifactTypeFile:
		file, err := BuildArtifactFile(ctx, a, opts)
		if err != nil {
			return err
		}

		return store.StoreFile(ctx, a, file)
	}

	return nil

}

func Command(r Registerer) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if err := Action(r, c); err != nil {
			return cli.Exit(err, 1)
		}
		return nil
	}
}

func BuildArtifactFile(ctx context.Context, a *pipeline.Artifact, opts *pipeline.ArtifactContainerOpts) (*dagger.File, error) {
	builder, err := a.Handler.Builder(ctx, opts)
	if err != nil {
		return nil, err
	}
	return a.Handler.BuildFile(ctx, builder, opts)
}

func BuildArtifactDirectory(ctx context.Context, a *pipeline.Artifact, opts *pipeline.ArtifactContainerOpts) (*dagger.Directory, error) {
	builder, err := a.Handler.Builder(ctx, opts)
	if err != nil {
		return nil, err
	}
	return a.Handler.BuildDir(ctx, builder, opts)
}

func ExportArtifactFunc(ctx context.Context, d *dagger.Client, sm *semaphore.Weighted, log *slog.Logger, v *pipeline.Artifact, store pipeline.ArtifactStore, dst string, checksum bool) func() error {
	return func() error {
		log.Info("Started exporting artifact...")

		log.Info("Acquiring semaphore")
		if err := sm.Acquire(ctx, 1); err != nil {
			log.Info("Error acquiring semaphore", "error", err)
			return err
		}
		log.Info("Acquired semaphore")

		defer sm.Release(1)

		filename, err := v.Handler.Filename(ctx)
		if err != nil {
			return fmt.Errorf("error processing artifact string '%s': %w", v.ArtifactString, err)
		}

		log.Info("Exporting artifact")
		paths, err := store.Export(ctx, d, v, dst, checksum)
		if err != nil {
			return fmt.Errorf("error exporting artifact '%s': %w", filename, err)
		}

		for _, v := range paths {
			fmt.Fprintf(Stdout, "%s\n", v)
		}

		log.Info("Done exporting artifact")

		return nil
	}
}
