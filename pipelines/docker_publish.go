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

	versionRepoBaseTags := make(map[string]map[string]map[BaseImage][]string)
	for i, v := range args.PackageInputOpts.Packages {
		base := BaseImageAlpine
		tarOpts := TarOptsFromFileName(v)
		targz := packages[i]

		if strings.Contains(v, "ubuntu") {
			base = BaseImageUbuntu
		}

		tags := GrafanaImageTags(base, args.DockerOpts.Registry, tarOpts)

		publisher = publisher.
			WithFile("/src/grafana.img", targz)

		for _, tag := range tags {
			publisher = publisher.
				WithExec([]string{"docker", "tag", "$(docker import grafana.img)", tag}).
				WithExec([]string{"docker", "push", tag})

			repo := strings.Split(tag, ":")[0]
			versionRepoBaseTags[tarOpts.Version][repo][base] = append(versionRepoBaseTags[tarOpts.Version][repo][base], tag)
		}
	}

	for version, repoBaseTags := range versionRepoBaseTags {
		for repo, baseTags := range repoBaseTags {
			for base, tags := range baseTags {
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
