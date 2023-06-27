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

// ValidatePackage downloads a package and validates from a Google Cloud Storage bucket.
func ValidatePackage(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return err
	}

	var (
		grp = &errgroup.Group{}
		sm  = semaphore.NewWeighted(args.ConcurrencyOpts.Parallel)
	)

	for i, file := range packages {
		name := args.PackageInputOpts.Packages[i]
		grp.Go(ValidatePackageFunc(ctx, sm, d, file, src, name))
	}

	return grp.Wait()
}

func ValidatePackageFunc(ctx context.Context, sm *semaphore.Weighted, d *dagger.Client, file *dagger.File, src *dagger.Directory, name string) func() error {
	return func() error {
		log.Printf("[%s] Attempting to validate package", name)
		log.Printf("[%s] Acquiring semaphore", name)
		if err := sm.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("failed to acquire semaphore: %w", err)
		}
		defer sm.Release(1)
		log.Printf("[%s] Acquired semaphore", name)

		log.Printf("[%s] Acquiring node version", name)
		nodeVersion, err := containers.NodeVersion(d, src).Stdout(ctx)
		if err != nil {
			return fmt.Errorf("failed to get node version from source code: %w", err)
		}

		log.Printf("[%s] Validating package", name)
		container, err := containers.ValidatePackage(d, file, src, nodeVersion)
		if err != nil {
			return fmt.Errorf("[%s] error: %w", name, err)
		}
		log.Printf("[%s] Done validating package", name)

		// TODO: Respect the --destination argument and format the sub-folder based on the package name
		if _, err := container.Directory("e2e/verify/specs").Export(ctx, "e2e-out"); err != nil {
			return err
		}

		// TODO: Print the directory name based on the --destination argument
		fmt.Fprintln(Stdout, "")
		return nil
	}
}
