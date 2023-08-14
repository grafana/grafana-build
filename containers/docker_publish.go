package containers

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

func PublishPackageImage(ctx context.Context, d *dagger.Client, pkg *dagger.File, tag string, opts *DockerOpts) (string, error) {
	return d.Container().From("docker").
		WithFile("grafana.img", pkg).
		WithUnixSocket("/var/run/docker.sock", d.Host().UnixSocket("/var/run/docker.sock")).
		WithExec([]string{"docker", "login", opts.Registry, "-u", opts.Username, "-p", opts.Password}).
		WithExec([]string{"/bin/sh", "-c", "docker load -i grafana.img | awk -F 'Loaded image: ' '{print $2}' > /tmp/image_tag"}).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("docker tag $(cat /tmp/image_tag) %s", tag)}).
		WithExec([]string{"docker", "push", tag}).
		Stdout(ctx)
}

func PublishDockerManifest(ctx context.Context, d *dagger.Client, manifest string, tags []string, opts *DockerOpts) (string, error) {
	return d.Container().From("docker").
		WithUnixSocket("/var/run/docker.sock", d.Host().UnixSocket("/var/run/docker.sock")).
		WithExec([]string{"docker", "login", opts.Registry, "-u", opts.Username, "-p", opts.Password}).
		WithExec(append([]string{"docker", "manifest", "create", manifest}, tags...)).
		WithExec([]string{"docker", "manifest", "push", manifest}).
		Stdout(ctx)
}
