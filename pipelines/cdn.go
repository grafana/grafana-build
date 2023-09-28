package pipelines

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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
			dst   = strings.Join([]string{args.PublishOpts.Destination, name, "public"}, "/")
		)

		log.Println("Copying frontend assets for", name, "to", dst)

		public := containers.ExtractedArchive(d, targz, name).Directory("public")
		dst, err := containers.PublishDirectory(ctx, d, public, args.GCPOpts, dst)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, dst)
	}
	return nil
}
