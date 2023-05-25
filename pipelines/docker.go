package pipelines

import (
	"context"
	"fmt"
	"log"
	"os"
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
func GrafanaImageTags(base BaseImage, registry string, opts TarFileOpts) []string {
	var (
		org     = "grafana"
		repos   = []string{"grafana-image-tags", "grafana-oss-image-tags"}
		version = opts.Version

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

func GetBaseImage(base BaseImage, distro executil.Distribution, opts *containers.DockerOpts) string {
	if base == BaseImageUbuntu {
		switch _, arch := executil.OSAndArch(distro); arch {
		case "arm64":
			return opts.UbuntuBaseARM64
		case "armv7":
			return opts.UbuntuBaseARMv7
		default:
			return opts.UbuntuBase
		}
	}
	switch _, arch := executil.OSAndArch(distro); arch {
	case "arm64":
		return opts.AlpineBaseARM64
	case "armv7":
		return opts.AlpineBaseARMv7
	default:
		return opts.AlpineBase
	}
}

// Docker is a pipeline that uses a grafana.tar.gz as input and creates a Docker image using that same Grafana's Dockerfile.
// Grafana's Dockerfile should support supplying a tar.gz using a --build-arg.
func Docker(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts)
	if err != nil {
		return err
	}
	var (
		opts        = args.DockerOpts
		publishOpts = args.PublishOpts
		saved       = map[string]*dagger.File{}
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
			var (
				platform  = executil.Platform(tarOpts.Distro)
				tags      = GrafanaImageTags(base, opts.Registry, tarOpts)
				baseImage = GetBaseImage(base, tarOpts.Distro, opts)
				socket    = d.Host().UnixSocket("/var/run/docker.sock")
			)

			log.Println("Building docker images", tags, "with base image", baseImage, "and platform", platform)

			args := []string{"GRAFANA_TGZ=grafana.tar.gz",
				fmt.Sprintf("BASE_IMAGE=%s", baseImage),
				"GO_SRC=tgz-builder",
				"JS_SRC=tgz-builder",
			}

			// Docker build and give the grafana.tar.gz as a build argument
			before := func(c *dagger.Container) *dagger.Container {
				return c.WithMountedFile("/src/Dockerfile", dockerfile).
					WithMountedFile("/src/packaging/docker/run.sh", runsh).
					WithMountedFile("/src/grafana.tar.gz", targz)
			}

			builder := containers.DockerBuild(d, &containers.DockerBuildOpts{
				Tags:       tags,
				BuildArgs:  args,
				UnixSocket: socket,
				Before:     before,
				Platform:   platform,
			})

			// if --save was provided then we will publish this to the requested location using PublishFile
			if publishOpts.Destination != "" {
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

	// The map of saved items will be nil if there was no destination set.
	// The images will still exist in the docker service though.
	for k, v := range saved {
		dst := strings.Join([]string{args.PublishOpts.Destination, k}, "/")
		out, err := containers.PublishFile(ctx, d, v, args.PublishOpts, dst)

		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, out)
	}

	return nil
}
