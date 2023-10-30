package containers

import (
	"context"

	"dagger.io/dagger"
)

func Run(ctx context.Context, containers []*dagger.Container) error {
	for _, v := range containers {
		if _, err := ExitError(ctx, v); err != nil {
			return err
		}
	}

	return nil
}
