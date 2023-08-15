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

		dst := strings.Join([]string{args.PublishOpts.Destination, name, "npm-artifacts"}, "/")

		log.Println("Copying npm artifacts for", name, "to", dst)
		folder := fmt.Sprintf("/src/%s", name)

		// gcloud rsync the artifacts folder to ['dst']/npm-artifacts
		artifacts := d.Container().From("busybox").
			WithFile("/src/grafana.tar.gz", targz).
			WithExec([]string{"mkdir", "-p", folder}).
			WithExec([]string{"tar", "--strip-components=1", "-xzf", "/src/grafana.tar.gz", "-C", folder}).
			Directory(folder + "/npm-artifacts")

		dst, err := containers.PublishDirectory(ctx, d, artifacts, args.GCPOpts, dst)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, dst)
	}
	return nil
}
