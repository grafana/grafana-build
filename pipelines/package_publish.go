package pipelines

import (
	"context"
	"log"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/slices"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// PublishPackage creates a package and publishes it to a Google Cloud Storage bucket.
func PublishPackage(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) error {
	var (
		opts = args.GrafanaOpts
		// This bool slice stores the values of args.BuildEnterprise for each build
		// --enterprise: []bool{true}
		// --enterprise --grafana: []bool{true, false}
		// if -- enterprise is not used it always returns []bool{false}
		skipOss           = opts.BuildEnterprise && !opts.BuildGrafana
		isEnterpriseBuild = slices.Unique([]bool{opts.BuildEnterprise, skipOss})
		distros           = args.PackageOpts.Distros
		nodeCache         = d.CacheVolume("yarn-dependencies")
	)

	files := map[string]*dagger.File{}

	for _, isEnterprise := range isEnterpriseBuild {
		var (
			src     = src
			edition = ""
		)
		if isEnterprise {
			edition = "enterprise"
			// If the user has manually set the edition flag, then override it with their selection.
			// Temporary fix: to avoid creating a --grafana and a --enterprise package with a conflicting name, only allow
			// overriding the 'edition' if we're building enterprise.
			if e := args.PackageOpts.Edition; e != "" {
				edition = e
			}
			s, err := args.GrafanaOpts.Enterprise(ctx, src, d)
			if err != nil {
				return err
			}
			src = s
		}

		opts := PackageOpts{
			GrafanaCompileOpts: &GrafanaCompileOpts{
				Source:   src,
				Version:  opts.Version,
				Platform: args.Platform,
				Env:      args.GrafanaOpts.Env,
				GoTags:   args.GrafanaOpts.GoTags,
			},
			BuildID:         opts.BuildID,
			Distributions:   distros,
			Edition:         edition,
			NodeCacheVolume: nodeCache,
		}
		packages, err := PackageFiles(ctx, d, opts)
		if err != nil {
			return err
		}

		for distro, targz := range packages {
			opts := TarFileOpts{
				Edition: opts.Edition,
				Version: opts.Version,
				BuildID: opts.BuildID,
				Distro:  distro,
			}

			fn := TarFilename(opts)
			dst := strings.Join([]string{args.PublishOpts.Destination, fn}, "/")
			log.Println("Writing package", fn, "to", dst)
			files[dst] = targz
		}
	}
	var (
		grp = &errgroup.Group{}
		sm  = semaphore.NewWeighted(int64(args.ConcurrencyOpts.Parallel))
	)

	for dst, file := range files {
		grp.Go(PublishFileFunc(ctx, sm, d, &containers.PublishFileOpts{
			File:        file,
			Destination: dst,
			PublishOpts: args.PublishOpts,
			GCPOpts:     args.GCPOpts,
		}))
	}

	return grp.Wait()
}
