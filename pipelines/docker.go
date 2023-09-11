package pipelines

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const (
	DefaultTagFormat       = "{{ .version }}-{{ .arch }}"
	DefaultUbuntuTagFormat = "{{ .version }}-ubuntu-{{ .arch }}"
)

type BaseImage int

const (
	BaseImageUbuntu BaseImage = iota
	BaseImageAlpine
)

type ImageTagOpts struct {
	Registry string
	Org      string
	Repo     string

	TarOpts TarFileOpts
}

func (i *ImageTagOpts) ValuesMap() map[string]string {
	arch := executil.FullArch(i.TarOpts.Distro)
	arch = strings.ReplaceAll(arch, "/", "")

	return map[string]string{
		"arch":    arch,
		"version": strings.TrimPrefix(i.TarOpts.Version, "v"),
		"buildID": i.TarOpts.BuildID,
	}
}

func ImageVersion(format string, values map[string]string) (string, error) {
	tmpl, err := template.New("version").Parse(format)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buf, values); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func ImageTag(format string, opts *ImageTagOpts) (string, error) {
	version, err := ImageVersion(format, opts.ValuesMap())
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s:%s", opts.Registry, opts.Org, opts.Repo, version), nil
}

// GrafanaImageTag returns the name of the grafana docker image based on the tar package name.
// To maintain backwards compatibility, we must keep this the same as it was before.
func GrafanaImageTags(base BaseImage, registry, tagFormat, ubuntuTagFormat string, opts TarFileOpts) ([]string, error) {
	var (
		org   = "grafana"
		repos = []string{"grafana-image-tags", "grafana-oss-image-tags"}

		edition = opts.Edition
	)

	if edition != "" {
		// Non-grafana repositories only create images in 1 repository instead of 2. Reason unknown.
		repos = []string{fmt.Sprintf("grafana-%s-image-tags", edition)}
	}

	tags := make([]string, len(repos))

	for i, repo := range repos {
		format := tagFormat
		if base == BaseImageUbuntu {
			format = ubuntuTagFormat
		}

		tag, err := ImageTag(format, &ImageTagOpts{
			Registry: registry,
			Org:      org,
			Repo:     repo,
			TarOpts:  opts,
		})
		if err != nil {
			return nil, err
		}

		tags[i] = tag
	}

	return tags, nil
}

// Docker is a pipeline that uses a grafana.tar.gz as input and creates a Docker image using that same Grafana's Dockerfile.
// Grafana's Dockerfile should support supplying a tar.gz using a --build-arg.
func Docker(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
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
			src        = containers.ExtractedArchive(d, targz, v)
			dockerfile = src.File("Dockerfile")
			runsh      = src.File("packaging/docker/run.sh")
		)

		bases := []BaseImage{BaseImageAlpine, BaseImageUbuntu}
		for _, base := range bases {
			var (
				platform  = executil.Platform(tarOpts.Distro)
				baseImage = opts.AlpineBase
				socket    = d.Host().UnixSocket("/var/run/docker.sock")
			)

			tags, err := GrafanaImageTags(base, opts.Registry, opts.TagFormat, opts.UbuntuTagFormat, tarOpts)
			if err != nil {
				return err
			}

			if base == BaseImageUbuntu {
				baseImage = opts.UbuntuBase
			}

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
				name := ReplaceExt(v, ext)
				img := builder.WithExec([]string{"docker", "save", tags[0], "-o", name}).File(name)
				dst := strings.Join([]string{publishOpts.Destination, name}, "/")
				saved[dst] = img
			}
		}
	}

	var (
		wg = &errgroup.Group{}
		sm = semaphore.NewWeighted(args.ConcurrencyOpts.Parallel)
	)
	for dst, file := range saved {
		wg.Go(PublishFileFunc(ctx, sm, d, &containers.PublishFileOpts{
			Destination: dst,
			File:        file,
			GCPOpts:     args.GCPOpts,
			PublishOpts: args.PublishOpts,
		}))
	}

	return wg.Wait()
}
