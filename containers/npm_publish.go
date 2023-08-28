package containers

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
)

// PublishNPM publishes a npm package to the given destination.
func PublishNPM(ctx context.Context, d *dagger.Client, pkg *dagger.File, opts *NPMOpts) (string, error) {
	src := ExtractedArchive(d, pkg, "pkg.tgz")

	version, err := GetJSONValue(ctx, d, src, "package.json", "version")
	if err != nil {
		return "", err
	}

	name, err := GetJSONValue(ctx, d, src, "package.json", "name")
	if err != nil {
		return "", err
	}

	major := strings.Split(version, ".")[0]
	minor := strings.Split(version, ".")[1]
	tag := fmt.Sprintf("latest-%s.%s", major, minor)

	c := d.Container().From(NodeImage("lts")).
		WithFile("/pkg.tgz", pkg).
		WithExec([]string{"npm", "set", fmt.Sprintf("//%s/:_authToken", opts.Registry), opts.Token}).
		WithExec([]string{"npm", "publish", "/pkg.tgz", fmt.Sprintf("--registry https://%s", opts.Registry), "--tag", tag})

	if opts.Latest {
		c = c.WithExec([]string{"npm", "dist-tag", "add", fmt.Sprintf("%s@%s", name, version), "latest"})
	}

	if opts.Next {
		c = c.WithExec([]string{"npm", "dist-tag", "add", fmt.Sprintf("%s@%s", name, version), "next"})
	}

	return c.Stdout(ctx)
}
