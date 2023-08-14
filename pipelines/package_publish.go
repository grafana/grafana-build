package pipelines

import (
	"context"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// PublishPackage takes one or multiple grafana.tar.gz as input and publishes it to a set destination.
func PublishPackage(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	var (
		wg = &errgroup.Group{}
		sm = semaphore.NewWeighted(args.ConcurrencyOpts.Parallel)
	)

	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return err
	}

	for i, name := range args.PackageInputOpts.Packages {
		wg.Go(PublishFileFunc(ctx, sm, d, &containers.PublishFileOpts{
			File:        packages[i],
			Destination: strings.Join([]string{args.PublishOpts.Destination, filepath.Base(name)}, "/"),
			PublishOpts: args.PublishOpts,
			GCPOpts:     args.GCPOpts,
		}))
	}

	return wg.Wait()
}
