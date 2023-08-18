package containers

import (
	"context"
	"fmt"
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

// GetLatestVersion gets the "version" field from https://grafana.com/api/grafana/versions/<channel>.
func GetLatestVersion(ctx context.Context, d *dagger.Client, channel string) (string, error) {
	c := d.Container().From("alpine").
		WithExec([]string{"apk", "--update", "add", "jq"}).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("curl https://grafana.com/api/grafana/versions/%s | jq -r .version", channel)})

	if stdout, err := c.Stdout(ctx); err == nil {
		return strings.TrimSpace(stdout), nil
	}

	return c.Stderr(ctx)
}

// IsLatest compares versions and returns true if the version provided is grater or equal the latest version on the channel.
func IsLatest(ctx context.Context, d *dagger.Client, channel string, version string) (bool, error) {
	latest, err := GetLatestVersion(ctx, d, channel)
	if err != nil {
		return false, err
	}

	c := d.Container().From("alpine").
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("echo -e '%s\n%s' | sort -V | tail -1", latest, version)})

	if stdout, err := c.Stdout(ctx); err == nil {
		return strings.TrimSpace(stdout) == version, nil
	}

	return false, nil
}
