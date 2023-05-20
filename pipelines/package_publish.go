package pipelines

import (
	"context"
	"log"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/slices"
)

// PublishPackage creates a package and publishes it to a Google Cloud Storage bucket.
func PublishPackage(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) error {
	opts := args.GrafanaOpts
	// This bool slice stores the values of args.BuildEnterprise for each build
	// --enterprise: []bool{true}
	// --enterprise --grafana: []bool{true, false}
	// if -- enterprise is not used it always returns []bool{false}
	skipOss := opts.BuildEnterprise && !opts.BuildGrafana
	isEnterpriseBuild := slices.Unique([]bool{opts.BuildEnterprise, skipOss})
	distros := args.PackageOpts.Distros
	for _, isEnterprise := range isEnterpriseBuild {
		edition := ""
		if isEnterprise {
			edition = "enterprise"
			// If the user has manually set the edition flag, then override it with their selection.
			// Temporary fix: to avoid creating a --grafana and a --enterprise package with a conflicting name, only allow
			// overriding the 'edition' if we're building enterprise.
			if e := args.PackageOpts.Edition; e != "" {
				edition = e
			}
		}

		opts := PackageOpts{
			GrafanaCompileOpts: &GrafanaCompileOpts{
				Source:   src,
				Version:  opts.Version,
				Platform: args.PackageOpts.Platform,
				Env:      args.GrafanaOpts.Env,
				GoTags:   args.GrafanaOpts.GoTags,
			},
			BuildID:       opts.BuildID,
			Distributions: distros,
			Edition:       edition,
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
			if err := containers.PublishFile(ctx, d, targz, args.PublishOpts, dst); err != nil {
				return err
			}
		}
	}

	return nil
}
