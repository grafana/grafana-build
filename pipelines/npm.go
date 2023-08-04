package pipelines

import (
	"context"
	"errors"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

// NPM publishes the NPM packages in the grafana.tar.gz(s)
func NPM(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return err
	}

	// Maybe we should allow this? not sure.
	if len(packages) != 1 {
		return errors.New("can not publish NPM packages for more than 1 tar.gz package")
	}

	var (
		targz  = packages[0]
		pkgdir = containers.ExtractedArchive(d, targz, args.PackageInputOpts.Packages[0])
	)

	c := containers.NodeContainer(d, "18.12.0", args.Platform).WithMountedDirectory("/src", pkgdir)
	return containers.ExitError(ctx, c)
}
