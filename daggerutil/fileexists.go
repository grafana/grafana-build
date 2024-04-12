package daggerutil

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

func FileExists(ctx context.Context, dir *dagger.Directory, path string) bool {
	_, err := dir.File(path).Contents(ctx)
	if err != nil {
		fmt.Printf("error checking if file exists: %s", err.Error())
		return false
	}

	return true
}
