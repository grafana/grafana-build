package frontend

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

// NPMPackages returns a dagger.Directory which contains the Grafana NPM packages from the grafana source code.
func NPMPackages(builder *dagger.Container, src *dagger.Directory, version string) *dagger.Directory {
	ersion := strings.TrimPrefix(version, "v")

	return builder.WithExec([]string{"mkdir", "npm-packages"}).
		WithExec([]string{"yarn", "run", "packages:build"}).
		// TODO: We should probably start reusing the yarn pnp map if we can figure that out instead of rerunning yarn install everywhere.
		WithExec([]string{"yarn", "run", "lerna", "version", ersion, "--exact", "--no-git-tag-version", "--no-push", "--force-publish", "-y"}).
		WithExec([]string{"yarn", "lerna", "exec", "--no-private", "--", "yarn", "pack", "--out", fmt.Sprintf("/src/npm-packages/%%s-%v.tgz", version)}).
		Directory("./npm-packages")
}

// PublishNPM publishes a npm package to the given destination.
func PublishNPM(ctx context.Context, d *dagger.Client, pkg *dagger.File, token, registry string, tags []string) (string, error) {
	src := containers.ExtractedArchive(d, pkg)

	version, err := containers.GetJSONValue(ctx, d, src, "package.json", "version")
	if err != nil {
		return "", err
	}

	name, err := containers.GetJSONValue(ctx, d, src, "package.json", "name")
	if err != nil {
		return "", err
	}

	tokenSecret := d.SetSecret("npm-token", token)

	c := d.Container().From(NodeImage("lts")).
		WithFile("/pkg.tgz", pkg).
		WithSecretVariable("NPM_TOKEN", tokenSecret).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("npm set //%s/:_authToken $NPM_TOKEN", registry)}).
		WithExec([]string{"npm", "publish", "/pkg.tgz", fmt.Sprintf("--registry https://%s", registry), "--tag", tags[0]})

	for _, tag := range tags[1:] {
		c = c.WithExec([]string{"npm", "dist-tag", "add", fmt.Sprintf("%s@%s", name, version), tag})
	}

	return c.Stdout(ctx)
}
