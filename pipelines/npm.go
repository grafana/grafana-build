package pipelines

import (
	"context"
	"fmt"
	"log"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

func NPM(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
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

		log.Println("Copying npm artifacts for", name, "to", args.PublishOpts.Destination)

		artifacts := containers.ExtractedArchive(d, targz, name).Directory("npm-artifacts")
		dst, err := containers.PublishDirectory(ctx, d, artifacts, args.GCPOpts, args.PublishOpts.Destination)
		if err != nil {
			return err
		}
		entries, err := artifacts.Entries(ctx)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			fmt.Fprintln(os.Stdout, dst+"/"+entry)
		}
	}
	return nil
}
