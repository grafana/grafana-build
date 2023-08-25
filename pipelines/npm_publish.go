package pipelines

import (
	"context"
	"fmt"
	"log"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

func PublishNPM(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	var (
		opts = args.NPMOpts
		wg   = &errgroup.Group{}
		sm   = semaphore.NewWeighted(args.ConcurrencyOpts.Parallel)
	)

	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return err
	}

	// Extract the package(s)
	for i, v := range args.PackageInputOpts.Packages {
		var (
			name  = ReplaceExt(v, "")
			targz = packages[i]
		)

		artifacts := containers.ExtractedArchive(d, targz, name).Directory("npm-artifacts")

		entries, err := artifacts.Entries(ctx)
		if err != nil {
			return err
		}

		for _, path := range entries {
			wg.Go(PublishNPMFunc(ctx, sm, d, artifacts.File(path), path, opts))
		}
	}
	return wg.Wait()
}

func PublishNPMFunc(ctx context.Context, sm *semaphore.Weighted, d *dagger.Client, pkg *dagger.File, path string, opts *containers.NPMOpts) func() error {
	return func() error {
		log.Printf("[%s] Attempting to publish package", path)
		log.Printf("[%s] Acquiring semaphore", path)
		if err := sm.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("failed to acquire semaphore: %w", err)
		}
		defer sm.Release(1)
		log.Printf("[%s] Acquired semaphore", path)

		log.Printf("[%s] Publishing package", path)
		out, err := containers.PublishNPM(ctx, d, pkg, opts)
		if err != nil {
			return fmt.Errorf("[%s] error: %w", path, err)
		}
		log.Printf("[%s] Done publishing package", path)

		fmt.Fprintln(Stdout, out)
		return nil
	}
}
