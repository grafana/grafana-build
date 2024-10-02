package frontend

import (
	"context"
	"fmt"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

// NPMPackages returns a dagger.Directory which contains the Grafana NPM packages from the grafana source code.
func NPMPackages(builder *dagger.Container, src *dagger.Directory, ersion string) *dagger.Directory {
	return builder.WithExec([]string{"mkdir", "npm-packages"}).
		WithEnvVariable("SHELL", "/bin/bash").
		WithExec([]string{"yarn", "install", "--immutable"}).
		WithExec([]string{"yarn", "run", "packages:build"}).
		WithExec([]string{"/bin/bash", "-c", fmt.Sprintf("yarn run lerna version %s --exact --no-git-tag-version --no-push --force-publish -y && yarn lerna exec --no-private -- yarn pack --out /src/npm-packages/%%s-%v.tgz", ersion, "v"+ersion)}).
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
