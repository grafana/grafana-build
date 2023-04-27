package pipelines

import (
	"context"
	"log"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
	"github.com/grafana/grafana-build/slices"
)

// PublishPackage creates a package and publishes it to a Google Cloud Storage bucket.
func PublishPackage(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) error {
	// This bool slice stores the values of args.BuildEnterprise for each build
	// --enterprise: []bool{true}
	// --enterprise --grafana: []bool{true, false}
	// if -- enterprise is not used it always returns []bool{false}
	skipOss := args.BuildEnterprise && !args.BuildGrafana
	isEnterpriseBuild := slices.Unique([]bool{args.BuildEnterprise, skipOss})
	distros := executil.DistrosFromStringSlice(args.Context.StringSlice("distro"))
	for _, isEnterprise := range isEnterpriseBuild {
		opts := PackageOpts{
			Src:          src,
			Version:      args.Version,
			BuildID:      args.BuildID,
			Distros:      distros,
			IsEnterprise: isEnterprise,
		}
		packages, err := PackageFiles(ctx, d, opts)
		if err != nil {
			return err
		}

		for distro, targz := range packages {
			opts := TarFileOpts{
				IsEnterprise: isEnterprise,
				Version:      args.Version,
				BuildID:      args.BuildID,
				Distro:       distro,
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
