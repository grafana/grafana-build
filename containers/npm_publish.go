package containers

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/executil"
)

// PublishNPM publishes a npm package to the given destination.
func PublishNPM(ctx context.Context, d *dagger.Client, pkg *dagger.File, opts *NPMOpts) (string, error) {
	src := ExtractedArchive(d, pkg, "pkg.tgz")
	version, err := GetJSONValue(ctx, d, src, "package.json", "version")
	name, err := GetJSONValue(ctx, d, src, "package.json", "name")

	isLatestStable, err := IsLatestGrafana(ctx, d, executil.Stable, version)
	if err != nil {
		return "", err
	}

	isLatestPreview, err := IsLatestGrafana(ctx, d, executil.Preview, version)
	if err != nil {
		return "", err
	}

	latestStable, err := GetLatestGrafanaVersion(ctx, d, executil.Stable)
	if err != nil {
		return "", err
	}

	latestPreview, err := GetLatestGrafanaVersion(ctx, d, executil.Preview)
	if err != nil {
		return "", err
	}

	isLaterThanPreview, err := IsLaterThan(ctx, d, version, latestPreview)
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
		// Workaround for now (maybe unnecessary?): set a NAME environment variable so that we don't accidentally cache
		WithEnvVariable("NAME", name).
		WithExec([]string{"npm", "set", fmt.Sprintf("//%s/:_authToken", opts.Registry), opts.Token})

	out, err := c.WithExec([]string{"npm", "view", name, "versions"}).Stdout(ctx)
	if !strings.Contains(out, fmt.Sprintf("'%s'", version)) {
		// Publish only if this version is not published already
		c = c.WithExec([]string{"npm", "publish", "/pkg.tgz", fmt.Sprintf("--registry https://%s", opts.Registry), "--tag", tag})
	}

	if !isLatestStable {
		c = c.WithExec([]string{"npm", "dist-tag", "add", fmt.Sprintf("%s@%s", name, latestStable), "latest"})
	}

	if isLatestPreview || (isLatestStable && isLaterThanPreview) {
		c = c.WithExec([]string{"npm", "dist-tag", "add", fmt.Sprintf("%s@%s", name, version), "next"})
	}

	return c.Stdout(ctx)
}
