package pipelines

import (
	"context"
	"fmt"
	"os"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

func Zip(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return err
	}
	zips := make(map[string]*dagger.File, len(packages))

	// Extract the package(s)
	for i, v := range args.PackageInputOpts.Packages {
		var (
			name  = DestinationName(v, "zip")
			targz = packages[i]
		)

		// We can't use path.Join here because publishopts.Destination might have a URL scheme which will get santizied, and we can't use filepath.Join because Windows would use \\ filepath separators.
		// dst := strings.Join([]string{args.PublishOpts.Destination, name, "public"}, "/")

		// gcloud rsync the public folder to ['dst']/public
		zip := d.Container().From("alpine").
			WithExec([]string{"apk", "add", "--update", "zip", "tar"}).
			WithFile("/src/grafana.tar.gz", targz).
			WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("tar xzf /src/grafana.tar.gz && zip %s $(tar tf /src/grafana.tar.gz)", name)}).
			File(name)
		zips[name] = zip
	}

	for k, v := range zips {
		dst := strings.Join([]string{args.PublishOpts.Destination, k}, "/")
		out, err := containers.PublishFile(ctx, d, &containers.PublishFileOpts{
			File:        v,
			Destination: dst,
			PublishOpts: args.PublishOpts,
			GCPOpts:     args.GCPOpts,
		})
		if err != nil {
			return err
		}

		fmt.Fprintln(os.Stdout, strings.Join(out, "\n"))
	}
	return nil
}
