package pipelines

import (
	"context"
	"fmt"
	"log"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/docker"
	"golang.org/x/sync/semaphore"
)

func ImageManifest(tag string) string {
	manifest := strings.ReplaceAll(tag, "-image-tags", "")
	lastDash := strings.LastIndex(manifest, "-")
	return manifest[:lastDash]
}

func LatestManifest(tag string) string {
	suffix := ""
	if strings.Contains(tag, "ubuntu") {
		suffix = "-ubuntu"
	}

	manifest := strings.ReplaceAll(tag, "-image-tags", "")
	manifestImage := strings.Split(manifest, ":")[0]
	return strings.Join([]string{manifestImage, fmt.Sprintf("latest%s", suffix)}, ":")
}

// PublishDocker is a pipeline that uses a grafana.docker.tar.gz as input and publishes a Docker image to a container registry or repository.
// Grafana's Dockerfile should support supplying a tar.gz using a --build-arg.
func PublishDocker(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	return nil
}

func PublishDockerManifestFunc(ctx context.Context, sm *semaphore.Weighted, d *dagger.Client, manifest string, tags []string, opts *DockerOpts) func() error {
	return func() error {
		log.Printf("[%s] Attempting to publish manifest", manifest)
		log.Printf("[%s] Acquiring semaphore", manifest)
		if err := sm.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("failed to acquire semaphore: %w", err)
		}
		defer sm.Release(1)
		log.Printf("[%s] Acquired semaphore", manifest)

		log.Printf("[%s] Publishing manifest", manifest)
		out, err := docker.PublishManifest(ctx, d, manifest, tags, opts.Username, opts.Password, opts.Registry)
		if err != nil {
			return fmt.Errorf("[%s] error: %w", manifest, err)
		}
		log.Printf("[%s] Done publishing manifest", manifest)

		fmt.Fprintln(Stdout, out)
		return nil
	}
}
