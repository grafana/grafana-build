package frontend

import (
	"context"
	"fmt"
	"log/slog"

	"dagger.io/dagger"
	"github.com/Masterminds/semver"
	"github.com/grafana/grafana-build/containers"
)

// NPMPackages returns a dagger.Directory which contains the Grafana NPM packages from the grafana source code.
func NPMPackages(ctx context.Context, builder *dagger.Container, d *dagger.Client, log *slog.Logger, src *dagger.Directory, ersion string) (*dagger.Directory, error) {
	lerna, err := containers.GetJSONValue(ctx, d, src, "package.json", "devDependencies.lerna")
	if err != nil {
		log.Error("Tried to read lerna version from package.json but failed.", "error", err)
		return nil, err
	}

	_, err = semver.NewVersion(lerna)
	if err != nil {
		nx, err := containers.GetJSONValue(ctx, d, src, "package.json", "devDependencies.nx")
		if err != nil {
			log.Error("Tried to read nx version from package.json but failed.", "error", err)
			return nil, err
		}

		_, err = semver.NewVersion(nx)
		if err != nil {
			log.Error("Could not find a valid version for either lerna or nx in package.json.", "error", err)
			return nil, err
		}

		// No lerna version specified in package.json
		return builder.WithExec([]string{"mkdir", "npm-packages"}).
			WithExec([]string{"yarn", "packages:build"}).
			WithExec([]string{"yarn", "nx", "release", "version", ersion, "--no-git-tag", "--no-git-commit", "--group", "grafanaPackages"}).
			// TODO: Jack - figure out how to do this with nx.
			// WithExec([]string{"yarn", "lerna", "exec", "--no-private", "--", "yarn", "pack", "--out", fmt.Sprintf("/src/npm-packages/%%s-%v.tgz", "v"+ersion)}).
			Directory("./npm-packages"), nil
	}

	return builder.WithExec([]string{"mkdir", "npm-packages"}).
		WithExec([]string{"yarn", "packages:build"}).
		WithExec([]string{"yarn", "lerna", "version", ersion, "--exact", "--no-git-tag-version", "--no-push", "--force-publish", "-y"}).
		WithExec([]string{"yarn", "lerna", "exec", "--no-private", "--", "yarn", "pack", "--out", fmt.Sprintf("/src/npm-packages/%%s-%v.tgz", "v"+ersion)}).
		Directory("./npm-packages"), nil
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
