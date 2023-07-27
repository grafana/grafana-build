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

		// We can't use path.Join here because publishopts.Destination might have a URL scheme which will get santizied, and we can't use filepath.Join because Windows would use \\ filepath separators.
		dst := strings.Join([]string{args.PublishOpts.Destination, name, "storybook"}, "/")

		log.Println("Copying storybook assets for", name, "to", dst)
		folder := fmt.Sprintf("/src/%s", name)

		// gcloud rsync the storybook folder to ['dst']/storybook
		storybook := d.Container().From("busybox").
			WithFile("/src/grafana.tar.gz", targz).
			WithExec([]string{"mkdir", "-p", folder}).
			WithExec([]string{"tar", "--strip-components=1", "-xzf", "/src/grafana.tar.gz", "-C", folder}).
			Directory(folder + "/storybook")

		dst, err := containers.PublishDirectory(ctx, d, storybook, args.GCPOpts, dst)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, dst)
	}
	return nil
}
