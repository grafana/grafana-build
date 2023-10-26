package pipelines

import (
	"context"
	"fmt"
	"log"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

func Storybook(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
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

		log.Println("Copying storybook assets for", name, "to", args.PublishOpts.Destination)

		storybook := containers.ExtractedArchive(d, targz).Directory("storybook")
		dst, err := containers.PublishDirectory(ctx, d, storybook, args.GCPOpts, args.PublishOpts.Destination)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, dst)
	}
	return nil
}
