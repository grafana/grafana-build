package pipelines

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

func NPMPackageName(path string) (string, string) {
	filename := filepath.Base(path)
	name := WithoutExt(filename)
	parts := strings.Split(name, "-")
	packageName := strings.Join([]string{parts[0], parts[1]}, "/")
	packageVersion := strings.Join(parts[2:], "-")
	return packageName, packageVersion
}

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
		name, version := NPMPackageName(path)
		log.Printf("[%s@%s] Attempting to publish package", name, version)
		log.Printf("[%s@%s] Acquiring semaphore", name, version)
		if err := sm.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("failed to acquire semaphore: %w", err)
		}
		defer sm.Release(1)
		log.Printf("[%s@%s] Acquired semaphore", name, version)

		log.Printf("[%s@%s] Publishing package", name, version)
		out, err := containers.PublishNPM(ctx, d, pkg, name, version, opts)
		if err != nil {
			return fmt.Errorf("[%s] error: %w", name, err)
		}
		log.Printf("[%s@%s] Done publishing package", name, version)

		fmt.Fprintln(Stdout, out)
		return nil
	}
}
