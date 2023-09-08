package containers

import (
	"context"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/errorutil"
)

func Run(ctx context.Context, containers []*dagger.Container) error {
	for _, v := range containers {
		if _, err := errorutil.ExitError(ctx, v); err != nil {
			return err
		}
	}

	return nil
}
