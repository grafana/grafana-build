package pipelines

import (
	"context"
	"fmt"
	"log"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
)

type BaseImage int

const (
	BaseImageUbuntu BaseImage = iota
	BaseImageAlpine
)

func ImageTag(registry, org, repo, version string) string {
	return fmt.Sprintf("%s/%s/%s:%s", registry, org, repo, version)
}

// GrafanaImageTag returns the name of the grafana docker image based on the tar package name.
// To maintain backwards compatibility, we must keep this the same as it was before.
func GrafanaImageTags(base BaseImage, opts TarFileOpts) []string {
	var (
		registry = "docker.io"
		org      = "grafana"
		repos    = []string{"grafana-image-tags", "grafana-oss-image-tags"}
		version  = opts.Version

		edition = opts.Edition
	)

	if edition != "" {
		// Non-grafana repositories only create images in 1 repository instead of 2. Reason unknown.
		repos = []string{fmt.Sprintf("grafana-%s-image-tags", edition)}
	}

	// For some unknown reason, versions in docker hub do not have a 'v'.
	// I think this was something that was established a long time ago and just stuck.
	version = strings.TrimPrefix(version, "v")

	if base == BaseImageUbuntu {
		version += "-ubuntu"
	}

	if opts.Distro != "" && opts.Distro != "linux/amd64" {
		_, arch := executil.OSAndArch(opts.Distro)
		version += "-" + strings.ReplaceAll(arch, "/", "")
	}

	tags := make([]string, len(repos))
	for i, repo := range repos {
		tags[i] = ImageTag(registry, org, repo, version)
	}

	return tags
}

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

		var (
			targz      = packages[i]
			src        = containers.ExtractedArchive(d, targz)
			dockerfile = src.File("Dockerfile")
			runsh      = src.File("packaging/docker/run.sh")
		)

		bases := []BaseImage{BaseImageAlpine, BaseImageUbuntu}
		for _, base := range bases {
			tags := GrafanaImageTags(base, tarOpts)
			log.Println("Building docker images", tags)
			// a different base image is used for arm versions of Grafana
			baseImage := args.DockerOpts.AlpineBase
			if base == BaseImageUbuntu {
				baseImage = args.DockerOpts.UbuntuBase
			}

			socket := d.Host().UnixSocket("/var/run/docker.sock")

			args := []string{"docker", "buildx", "build", ".",
				"--build-arg=GRAFANA_TGZ=grafana.tar.gz",
				fmt.Sprintf("--build-arg=BASE_IMAGE=%s", baseImage),
				"--build-arg=GO_SRC=tgz-builder",
				"--build-arg=JS_SRC=tgz-builder",
			}

			for _, v := range tags {
				args = append(args, "-t", v)
			}

			// Docker build and give the grafana.tar.gz as a build argument
			builder := d.Container().From("docker").
				WithUnixSocket("/var/run/docker.sock", socket).
				WithWorkdir("/src").
				WithMountedFile("/src/Dockerfile", dockerfile).
				WithMountedFile("/src/packaging/docker/run.sh", runsh).
				WithMountedFile("/src/grafana.tar.gz", targz).
				WithExec(args)

			// if --save was provided then we will publish this to the requested location using PublishFile
			if opts.Save {
				ext := "docker.tar.gz"
				if base == BaseImageUbuntu {
					ext = "ubuntu.docker.tar.gz"
				}
				name := DestinationName(v, ext)
				img := builder.WithExec([]string{"docker", "save", tags[0], "-o", name}).File(name)
				saved[name] = img
			}
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
