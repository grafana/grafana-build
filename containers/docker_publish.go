package containers

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

func PublishPackageImage(ctx context.Context, d *dagger.Client, pkg *dagger.File, tag string, opts *DockerOpts) (string, error) {
	username := d.SetSecret("docker-username", opts.Username)
	password := d.SetSecret("docker-password", opts.Password)

	return d.Container().From("docker").
		WithFile("grafana.img", pkg).
		WithSecretVariable("DOCKER_USERNAME", username).
		WithSecretVariable("DOCKER_PASSWORD0", password).
		WithUnixSocket("/var/run/docker.sock", d.Host().UnixSocket("/var/run/docker.sock")).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("docker login %s -u $DOCKER_USERNAME -p $DOCKER_PASSWORD", opts.Registry)}).
		WithExec([]string{"/bin/sh", "-c", "docker load -i grafana.img | awk -F 'Loaded image: ' '{print $2}' > /tmp/image_tag"}).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("docker tag $(cat /tmp/image_tag) %s", tag)}).
		WithExec([]string{"docker", "push", tag}).
		Stdout(ctx)
}

func PublishDockerManifest(ctx context.Context, d *dagger.Client, manifest string, tags []string, opts *DockerOpts) (string, error) {
	username := d.SetSecret("docker-username", opts.Username)
	password := d.SetSecret("docker-password", opts.Password)

	return d.Container().From("docker").
		WithUnixSocket("/var/run/docker.sock", d.Host().UnixSocket("/var/run/docker.sock")).
		WithSecretVariable("DOCKER_USERNAME", username).
		WithSecretVariable("DOCKER_PASSWORD0", password).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("docker login %s -u $DOCKER_USERNAME -p $DOCKER_PASSWORD", opts.Registry)}).
		WithExec(append([]string{"docker", "manifest", "create", manifest}, tags...)).
		WithExec([]string{"docker", "manifest", "push", manifest}).
		Stdout(ctx)
}
