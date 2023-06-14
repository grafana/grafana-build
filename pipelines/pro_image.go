package pipelines

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

func ProImage(ctx context.Context, dc *dagger.Client, directory *dagger.Directory, args PipelineArgs) error {
	if len(args.PackageInputOpts.Packages) > 1 {
		return fmt.Errorf("only one package is allowed: packages=%+v", args.PackageInputOpts.Packages)
	}
	packages, err := containers.GetPackages(ctx, dc, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return fmt.Errorf("getting packages: packages=%+v %w", args.PackageInputOpts.Packages, err)
	}

	debianPackageFile := packages[0]

	hostedGrafanaRepo, err := containers.CloneWithGitHubToken(dc, args.ProImageOpts.GithubToken, "https://github.com/grafana/hosted-grafana.git", "main")
	if err != nil {
		return fmt.Errorf("cloning hosted-grafana repo: %w", err)
	}

	socketPath := "/var/run/docker.sock"
	socket := dc.Host().UnixSocket(socketPath)

	hostedGrafanaImageTag := fmt.Sprintf("hosted-grafana-pro:%s", args.ProImageOpts.ImageTag)

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
			return fmt.Errorf("--registry=<string> is required")
		}

		authenticator := containers.GCSAuth(dc, &containers.GCPOpts{
			ServiceAccountKey:       args.GCPOpts.ServiceAccountKey,
			ServiceAccountKeyBase64: args.GCPOpts.ServiceAccountKeyBase64,
		})

		publishContainer := dc.Container().From("google/cloud-sdk:433.0.0-alpine")

		authenticatedContainer, err := authenticator.Authenticate(dc, publishContainer)
		if err != nil {
			return fmt.Errorf("authenticating container with gcs auth: %w", err)
		}

		address := fmt.Sprintf("%s/%s", args.ProImageOpts.ContainerRegistry, hostedGrafanaImageTag)

		ref, err := authenticatedContainer.Publish(ctx, address)
		if err != nil {
			return fmt.Errorf("publishing container: address=%s %w", address, err)
		}

		n, err := fmt.Fprintln(os.Stdout, ref)
		if err != nil {
			return fmt.Errorf("writing ref to stdout: bytesWritten=%d %w", n, err)
		}
	}

	return nil
}
