package containers

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/executil"
)

// GetJSONValue gets the value of a JSON field from a JSON file in the 'src' directory.
func GetJSONValue(ctx context.Context, d *dagger.Client, src *dagger.Directory, file string, field string) (string, error) {
	c := d.Container().From("alpine").
		WithExec([]string{"apk", "--update", "add", "jq"}).
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("cat %s | jq -r .%s", file, field)})

	if stdout, err := c.Stdout(ctx); err == nil {
		return strings.TrimSpace(stdout), nil
	}

	return c.Stderr(ctx)
}

// GetLatestGrafanaVersion gets the "version" field from https://grafana.com/api/grafana/versions/<channel>.
func GetLatestGrafanaVersion(ctx context.Context, d *dagger.Client, channel executil.VersionChannel) (string, error) {
	c := d.Container().From("alpine").
		WithExec([]string{"apk", "--update", "add", "jq", "curl"}).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("curl -s https://grafana.com/api/grafana/versions/%s | jq -r .version", channel)})

	if stdout, err := c.Stdout(ctx); err == nil {
		out := strings.TrimSpace(stdout)
		if out == "" {
			return out, fmt.Errorf("failed to retrieve grafana version from grafana.com")
		}
		return out, nil
	}

	return c.Stderr(ctx)
}

// IsLatestGrafana compares versions and returns true if the version provided is grater or equal the latest version of Grafana on the channel.
func IsLatestGrafana(ctx context.Context, d *dagger.Client, channel executil.VersionChannel, version string) (bool, error) {
	if versionChannel := executil.GetVersionChannel(version); versionChannel != channel {
		return false, nil
	}

	latestGrafana, err := GetLatestGrafanaVersion(ctx, d, channel)
	if err != nil {
		return false, err
	}

	return IsLaterThan(ctx, d, version, latestGrafana)
}

// GetLatestVersion compares versions and returns the latest version provided in the slice.
func GetLatestVersion(ctx context.Context, d *dagger.Client, versions []string) (string, error) {
	c := d.Container().From("alpine").
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("echo -e '%s' | sort -V | tail -1", strings.Join(versions, "\\n"))})

	stdout, err := c.Stdout(ctx)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(stdout), nil
}

// IsLaterThan compares versions and returns true if v1 is later than v2
func IsLaterThan(ctx context.Context, d *dagger.Client, v1 string, v2 string) (bool, error) {
	latest, err := GetLatestVersion(ctx, d, []string{v1, v2})
	if err != nil {
		return false, err
	}
	return latest == v1, nil
}
