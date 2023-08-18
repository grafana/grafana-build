package containers

import (
	"context"
	"fmt"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/executil"
)

// PublishNPM publishes a npm package to the given destination.
func PublishNPM(ctx context.Context, d *dagger.Client, pkg *dagger.File, name string, version string, opts *NPMOpts) (string, error) {
	isLatestStable, err := IsLatest(ctx, d, executil.Stable, version)
	if err != nil {
		return "", err
	}

	isLatestPreview, err := IsLatest(ctx, d, executil.Preview, version)
	if err != nil {
		return "", err
	}

	latest, err := GetLatestVersion(ctx, d, executil.Stable)
	if err != nil {
		return "", err
	}

	tag := "latest"
	if channel := executil.GetVersionChannel(version); channel == executil.Test {
		tag = "test"
	}

	if channel := executil.GetVersionChannel(version); channel == executil.Nightly {
		tag = "canary"
	}

	c := d.Container().From(NodeImage("lts")).
		WithFile("/pkg.tgz", pkg).
		WithExec([]string{"npm", "set", fmt.Sprintf("//%s/:_authToken", opts.Registry), opts.Token}).
		WithExec([]string{"npm", "publish", "/pkg.tgz", "--registry", opts.Registry, "--tag", tag})

	if !isLatestStable {
		c = c.WithExec([]string{"npm", "dist-tag", "add", fmt.Sprintf("%s@%s", name, latest), "latest"})
	}

	if isLatestPreview {
		c = c.WithExec([]string{"npm", "dist-tag", "add", fmt.Sprintf("%s@%s", name, version), "next"})
	}

	return c.Stdout(ctx)
}
