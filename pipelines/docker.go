package pipelines

import (
	"context"
	"fmt"
	"log"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

// Docker is a pipeline that uses a grafana.tar.gz as input and creates a Docker image using that same Grafana's Dockerfile.
// Grafana's Dockerfile should support supplying a tar.gz using the
func Docker(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts)
	if err != nil {
		return err
	}

	var (
		opts  = args.DockerOpts
		saved = map[string]*dagger.File{}
	)

	for i, v := range args.PackageInputOpts.Packages {
		tarOpts := TarOptsFromFileName(v)
		name := "grafana"
		if edition := tarOpts.Edition; edition != "" {
			name = fmt.Sprintf("%s-%s", name, edition)
		}

		var (
			targz      = packages[i]
			tag        = fmt.Sprintf("%s:%s", name, tarOpts.Version)
			src        = containers.ExtractedArchive(d, targz)
			dockerfile = src.File("Dockerfile")
			runsh      = src.File("packaging/docker/run.sh")
		)

		socket := d.Host().UnixSocket("/var/run/docker.sock")
		// Docker build and give the grafana.tar.gz as a build argument
		builder := d.Container().From("docker").
			WithUnixSocket("/var/run/docker.sock", socket).
			WithWorkdir("/src").
			WithMountedFile("/src/Dockerfile", dockerfile).
			WithMountedFile("/src/packaging/docker/run.sh", runsh).
			WithMountedFile("/src/grafana.tar.gz", targz).
			WithExec([]string{"docker", "buildx", "build", ".", "--build-arg=GRAFANA_TGZ=/src/grafana.tar.gz", "-t", tag})

		// if --save was provided then we will publish this to the requested location using PublishFile
		if opts.Save {
			name := DestinationName(v, "img")
			img := builder.WithExec([]string{"docker", "save", tag, "-o", name}).File(name)
			saved[name] = img
		}
	}

	for k, v := range saved {
		dst := strings.Join([]string{args.PublishOpts.Destination, k}, "/")
		log.Println(k, v, dst)
		if err := containers.PublishFile(ctx, d, v, args.PublishOpts, dst); err != nil {
			return err
		}
	}

	return nil
}
