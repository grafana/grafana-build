package pipelines

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

// DockerPublish is a pipeline that uses a grafana.docker.tar.gz as input and publishes a Docker image to a container registry or repository.
// Grafana's Dockerfile should support supplying a tar.gz using a --build-arg.
func DockerPublish(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return err
	}

	socket := d.Host().UnixSocket("/var/run/docker.sock")
	publisher := d.Container().From("docker").
		WithUnixSocket("/var/run/docker.sock", socket)

	versionRepoTags := make(map[string]map[string][]string)
	bases := []BaseImage{BaseImageAlpine, BaseImageUbuntu}
	for _, base := range bases {
		for i, v := range args.PackageInputOpts.Packages {
			tarOpts := TarOptsFromFileName(v)
			targz := packages[i]

			tags := GrafanaImageTags(base, args.DockerOpts.Registry, tarOpts)

			publisher = publisher.
				WithFile("/src/grafana.img", targz)

			for _, tag := range tags {
				publisher = publisher.
					WithExec([]string{"docker", "tag", "$(docker import grafana.img)", tag}).
					WithExec([]string{"docker", "push", tag})

				repo := strings.Split(tag, ":")[0]
				versionRepoTags[tarOpts.Version][repo] = append(versionRepoTags[tarOpts.Version][repo], tag)
			}
		}

		for version, repoTags := range versionRepoTags {
			for repo, tags := range repoTags {
				version := strings.TrimPrefix(version, "v")
				if base == BaseImageUbuntu {
					version += "-ubuntu"
				}

				manifestRepo := strings.TrimSuffix(repo, "-image-tags")
				manifestTag := fmt.Sprintf("%s:%s", manifestRepo, version)
				publisher = publisher.
					WithExec(append([]string{"docker", "manifest", "create", manifestTag}, tags...)).
					WithExec([]string{"docker", "manifest", "push", manifestTag})
			}
		}
	}

	return nil
}
