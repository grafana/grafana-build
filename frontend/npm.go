package frontend

import (
	"context"
	"fmt"
	"log/slog"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/daggerutil"
)

// NPMPackages versions and packs the npm packages into tarballs into `npm-packages` directory.
// It then returns the npm-packages directory as a dagger.Directory.
func NPMPackages(ctx context.Context, builder *dagger.Container, log *slog.Logger, src *dagger.Directory, ersion string) (*dagger.Directory, error) {
	// Check if the version of Grafana uses lerna or nx to manage package versioning.
	hasLerna := daggerutil.FileExists(ctx, src, "./lerna.json")
	hasNx := daggerutil.FileExists(ctx, src, "./nx.json")

	if hasLerna {
		return builder.WithExec([]string{"mkdir", "npm-packages"}).
			WithExec([]string{"yarn", "packages:build"}).
			WithExec([]string{"yarn", "lerna", "version", ersion, "--exact", "--no-git-tag-version", "--no-push", "--force-publish", "-y"}).
			WithExec([]string{"yarn", "lerna", "exec", "--no-private", "--", "yarn", "pack", "--out", fmt.Sprintf("/src/npm-packages/%%s-%v.tgz", "v"+ersion)}).
			Directory("./npm-packages"), nil
	}

	if hasNx {
		return builder.WithExec([]string{"mkdir", "npm-packages"}).
			WithExec([]string{"yarn", "packages:build"}).
			WithExec([]string{"yarn", "nx", "release", "version", ersion, "--no-git-tag", "--no-git-commit", "--no-stage-changes", "--group", "grafanaPackages"}).
			WithExec([]string{"yarn", "workspaces", "foreach", "--no-private", "--include='@grafana/*'", "-A", "exec", "yarn", "pack", "--out", fmt.Sprintf("/src/npm-packages/%%s-%v.tgz", "v"+ersion)}).
			Directory("./npm-packages"), nil
	}

	log.Error("No lerna.json or nx.json found in source directory")
	return nil, fmt.Errorf("no lerna.json or nx.json found in source directory")
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
