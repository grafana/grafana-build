package pipelines

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

// RPMPublish uses the grafana rpm given by the '--package' argument and publishes it to our yum repository.
func RPMPublish(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	c, err := containers.PackagePublish(ctx, d, args.PackagePublishOpts, "rpm")

	if err != nil {
		return err
	}

	out, err := c.Stdout(ctx)
	if err != nil {
		return err
	}

	fmt.Fprint(os.Stdout, out)
	return nil
}
