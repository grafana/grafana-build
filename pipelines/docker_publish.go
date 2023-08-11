package pipelines

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

func ImageManifest(tag string) string {
	manifest := strings.ReplaceAll(tag, "-image-tags", "")
	lastDash := strings.LastIndex(manifest, "-")
	return manifest[:lastDash]
}

// DockerPublish is a pipeline that uses a grafana.docker.tar.gz as input and publishes a Docker image to a container registry or repository.
// Grafana's Dockerfile should support supplying a tar.gz using a --build-arg.
func DockerPublish(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return err
	}

	socket := d.Host().UnixSocket("/var/run/docker.sock")
	publisher := d.Container().From("docker").
		WithUnixSocket("/var/run/docker.sock", socket).
		WithExec([]string{"docker", "login", args.DockerOpts.Registry, "-u", args.DockerOpts.Username, "-p", args.DockerOpts.Password})

	manifestTags := make(map[string][]string)
	for i, v := range args.PackageInputOpts.Packages {
		base := BaseImageAlpine
		tarOpts := TarOptsFromFileName(v)
		targz := packages[i]

		if strings.Contains(v, "ubuntu") {
			base = BaseImageUbuntu
		}

		tags := GrafanaImageTags(base, args.DockerOpts.Registry, tarOpts)

		publisher = publisher.
			WithFile("grafana.img", targz)

		for _, tag := range tags {
			publisher = publisher.
				WithExec([]string{"/bin/sh", "-c", "docker load -i grafana.img | awk -F 'Loaded image: ' '{print $2}' > /tmp/image_tag"}).
				WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("docker tag $(cat /tmp/image_tag) %s", tag)}).
				WithExec([]string{"docker", "push", tag})

			manifest := ImageManifest(tag)
			manifestTags[manifest] = append(manifestTags[manifest], tag)
		}
	}

	for manifest, tags := range manifestTags {
		publisher = publisher.
			WithExec(append([]string{"docker", "manifest", "create", manifest}, tags...)).
			WithExec([]string{"docker", "manifest", "push", manifest})
	}

	_, err = publisher.ExitCode(ctx)
	return err
}
