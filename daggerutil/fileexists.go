package daggerutil

import (
	"context"

	"dagger.io/dagger"
)

func FileExists(ctx context.Context, dir *dagger.Directory, path string) bool {
	_, err := dir.File(path).Contents(ctx)
	if err == nil {
		return true
	}

	return false
}
