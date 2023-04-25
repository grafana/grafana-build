package pipelines

import (
	"context"
	"log"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

// PublishPackage creates a package and publishes it to a Google Cloud Storage bucket.
func PublishPackage(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) error {
	packages, err := PackageFiles(ctx, d, src, args)
	if err != nil {
		return err
	}

	for distro, targz := range packages {
		fn := TarFilename(args.Version, args.BuildID, args.BuildEnterprise, distro)
		dst := strings.Join([]string{args.Context.Path("destination"), fn}, "/")
		log.Println("Writing package", fn, "to", dst)
		if err := containers.PublishFile(ctx, d, targz, args.PublishOpts, dst); err != nil {
			return err
		}
	}

	return nil
}
