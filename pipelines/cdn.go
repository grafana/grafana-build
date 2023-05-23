package pipelines

import (
	"context"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

func CDN(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts)
	if err != nil {
		return err
	}

	// Extract the package(s)
	for i, v := range args.PackageInputOpts.Packages {
		var (
			name  = DestinationName(v, "")
			targz = packages[i]
		)
		// gcloud rsync the public folder to ['dst']/public
		public := d.Container().From("busybox").
			WithFile("/src/grafana.tar.gz", targz).
			WithExec([]string{"mkdir", "-p", "/src/grafana"}).
			WithExec([]string{"tar", "--strip-components=1", "-xzf", "/src/grafana.tar.gz", "-C", "/src/grafana"}).
			WithWorkdir("/src").
			Directory("/src/grafana/public")

			// We can't use path.Join here because publishopts.Destination might have a URL scheme which will get santizied, and we can't use filepath.Join because Windows would use \\ filepath separators.
		dst := strings.Join([]string{args.PublishOpts.Destination, name, "public"}, "/")

		if err := containers.PublishDirectory(ctx, d, public, args.PublishOpts, dst); err != nil {
			return err
		}
	}
	return nil
}
