package pipelines

import (
	"context"
	"fmt"
	"log"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

func CDN(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
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

		log.Println("Copying frontend assets for", name, "to", args.PublishOpts.Destination)

		public := containers.ExtractedArchive(d, targz, name).Directory("public")
		dst, err := containers.PublishDirectory(ctx, d, public, args.GCPOpts, args.PublishOpts.Destination)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, dst)
	}
	return nil
}
