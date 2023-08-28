package containers

import (
	"context"
	"fmt"

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

	token := d.SetSecret("npm-token", opts.Token)

	c := d.Container().From(NodeImage("lts")).
		WithFile("/pkg.tgz", pkg).
		WithSecretVariable("NPM_TOKEN", token).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("npm set //%s/:_authToken $NPM_TOKEN", opts.Registry)}).
		WithExec([]string{"npm", "publish", "/pkg.tgz", fmt.Sprintf("--registry https://%s", opts.Registry), "--tag", opts.Tags[0]})

	for _, tag := range opts.Tags[1:] {
		c = c.WithExec([]string{"npm", "dist-tag", "add", fmt.Sprintf("%s@%s", name, version), tag})
	}

	return c.Stdout(ctx)
}
