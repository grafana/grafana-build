package pipelines

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

// PublishPackage takes one or multiple grafana.tar.gz as input and publishes it to a set destination.
func PublishPackage(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return err
	}

	c := d.Container().From("alpine")
	for i, name := range args.PackageInputOpts.Packages {
		c = c.WithFile("/dist/"+filepath.Base(name), packages[i])
	}

	dst, err := containers.PublishDirectory(ctx, d, c.Directory("dist"), args.GCPOpts, args.PublishOpts.Destination)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, dst)
	return nil
}
