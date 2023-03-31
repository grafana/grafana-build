package containers

import (
	"context"
	"strings"

	"dagger.io/dagger"
)

// GetPackageJSONVersion gets the "version" field from package.json in the 'src' directory.
func GetPackageJSONVersion(ctx context.Context, d *dagger.Client, src *dagger.Directory) (string, error) {
	c := d.Container().From("alpine").
		WithExec([]string{"apk", "--update", "add", "jq"}).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"/bin/sh", "-c", "cat package.json | jq -r .version"})

	if stdout, err := c.Stdout(ctx); err == nil {
		return strings.TrimSpace(stdout), nil
	}

	return c.Stderr(ctx)
}
