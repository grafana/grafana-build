package pipelines

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/fpm"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type InstallerFunc func(ctx context.Context, d *dagger.Client, args PipelineArgs, opts fpm.BuildOpts) error

func PublishInstallers(ctx context.Context, d *dagger.Client, args PipelineArgs, packages map[string]*dagger.File) error {
	var (
		wg = &errgroup.Group{}
		sm = semaphore.NewWeighted(args.ConcurrencyOpts.Parallel)
	)

	for dst, file := range packages {
		wg.Go(PublishFileFunc(ctx, sm, d, &containers.PublishFileOpts{
			Destination: dst,
			File:        file,
			GCPOpts:     args.GCPOpts,
			PublishOpts: args.PublishOpts,
		}))
	}

	return wg.Wait()
}

// Uses the grafana package given by the '--package' argument and creates a installer.
// It accepts publish args, so you can place the file in a local or remote destination.
func PackageInstaller(ctx context.Context, d *dagger.Client, args PipelineArgs, opts fpm.BuildOpts) (map[string]*dagger.File, error) {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return nil, err
	}

	installers := make(map[string]*dagger.File, len(packages))

	for i, v := range args.PackageInputOpts.Packages {
		name := filepath.Base(strings.TrimPrefix(strings.ReplaceAll(v, ".tar.gz", fmt.Sprintf(".%s", opts.PackageType)), "file://"))
		packageName := string(opts.Name)
		if n := opts.NameOverride; n != "" {
			packageName = n
		}
		dst := strings.Join([]string{args.PublishOpts.Destination, strings.ReplaceAll(name, string(opts.Name), packageName)}, "/")
		installers[dst] = fpm.Build(fpm.Builder(d), opts, packages[i])
	}

	return installers, nil
}
