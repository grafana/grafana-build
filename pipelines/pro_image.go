package pipelines

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

func ProImage(ctx context.Context, dc *dagger.Client, directory *dagger.Directory, args PipelineArgs) error {
	debianPackageFile := dc.Host().Directory("./").File(args.ProImageOpts.Deb)

	hostedGrafanaRepo, err := containers.CloneWithGitHubToken(dc, args.ProImageOpts.GithubToken, "https://github.com/grafana/hosted-grafana.git", "main")
	if err != nil {
		return fmt.Errorf("cloning hosted-grafana repo: %w", err)
	}

	socketPath := "/var/run/docker.sock"
	socket := dc.Host().UnixSocket(socketPath)

	hostedGrafanaImageTag := fmt.Sprintf("hosted-grafana-pro:%s", args.ProImageOpts.GrafanaVersion)

	container := dc.Container().From("golang:1.20-alpine").
		WithUnixSocket(socketPath, socket).
		WithExec([]string{"apk", "add", "--update", "docker"}).
		WithDirectory("/src", hostedGrafanaRepo).
		WithFile("/src/grafana.deb", debianPackageFile).
		WithWorkdir("/src").
		WithExec([]string{
			"/bin/sh", "-c",
			"docker build --platform=linux/amd64 . -f ./cmd/hgrun/Dockerfile -t hgrun:latest",
		}).
		WithExec([]string{
			"/bin/sh", "-c",
			fmt.Sprintf("docker build --platform=linux/amd64 --build-arg=RELEASE_TYPE=%s --build-arg=GRAFANA_VERSION=%s --build-arg=HGRUN_IMAGE=hgrun:latest . -f ./docker/hosted-grafana-all/Dockerfile -t %s",
				args.ProImageOpts.ReleaseType,
				args.ProImageOpts.GrafanaVersion,
				hostedGrafanaImageTag,
			),
		})

	if err := containers.ExitError(ctx, container); err != nil {
		return fmt.Errorf("container did not exit successfully: %w", err)
	}

	if args.ProImageOpts.Push {
		if args.ProImageOpts.ContainerRegistry == "" {
			return fmt.Errorf("--container-registry=<string> is required")
		}

		publishContainer := dc.Container().From("google/cloud-sdk:alpine")

		authenticator := containers.GCSAuth(dc, &containers.GCPOpts{
			ServiceAccountKey:       args.GCPOpts.ServiceAccountKey,
			ServiceAccountKeyBase64: args.GCPOpts.ServiceAccountKeyBase64,
		})

		authenticatedContainer, err := authenticator.Authenticate(dc, publishContainer)
		if err != nil {
			return fmt.Errorf("authenticating container with gcs auth: %w", err)
		}

		image := fmt.Sprintf("%s/%s", args.ProImageOpts.ContainerRegistry, hostedGrafanaImageTag)

		ref, err := authenticatedContainer.Publish(ctx, image)
		if err != nil {
			return fmt.Errorf("publishing container: %w", err)
		}

		fmt.Fprintln(os.Stdout, ref)
	}

	return nil
}
