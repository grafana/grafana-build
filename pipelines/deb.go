package pipelines

import (
	"context"

	"dagger.io/dagger"
)

// Deb uses the grafana package given by the '--package' argument and creates a .deb installer.
// It accepts publish args, so you can place the file in a local or remote destination.
func Deb(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) error {
	return nil
}
