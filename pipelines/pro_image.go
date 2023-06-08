package pipelines

import (
	"context"
	"fmt"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

func ProImage(ctx context.Context, dc *dagger.Client, directory *dagger.Directory, args PipelineArgs) error {
	debianPackageFile := dc.Host().Directory("./").File(args.ProImageOpts.Deb)

	hostedGrafanaRepo, err := containers.CloneWithGitHubToken(dc, args.ProImageOpts.GithubToken, "https://github.com/grafana/hosted-grafana", "main")
	if err != nil {
		return fmt.Errorf("cloning hosted-grafana repo: %w", err)
	}

	socketPath := "/var/run/docker.sock"
	socket := dc.Host().UnixSocket(socketPath)

	container := dc.Container().From("golang:1.20-alpine").
		WithUnixSocket(socketPath, socket).
		WithExec([]string{"apk", "add", "--update", "make", "docker", "git", "jq", "jsonnet"}).
		// WithExec([]string{"curl", "-O", "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-434.0.0-linux-x86_64.tar.gz"}).
		// WithExec([]string{"tar", "-xf", "google-cloud-cli-434.0.0-linux-x86_64.tar.gz", "--strip-components=1", "-C", "/"}).
		// WithExec([]string{"/install.sh", "--disable-prompts"}).
		WithDirectory("/src", hostedGrafanaRepo).
		WithFile("/src/grafana.deb", debianPackageFile, dagger.ContainerWithFileOpts{}).
		WithWorkdir("/src").
		WithExec([]string{
			"/bin/sh", "-c",
			"docker build --platform=linux/amd64 . -f ./cmd/hgrun/Dockerfile -t hgrun:latest",
		}).
		WithExec([]string{
			"/bin/sh", "-c",
			fmt.Sprintf("docker build --platform=linux/amd64 --build-arg=RELEASE_TYPE=%s --build-arg=GRAFANA_VERSION=%s --build-arg=HGRUN_IMAGE=hgrun:latest . -f ./docker/hosted-grafana-all/Dockerfile -t hostedgrafana:latest", args.ProImageOpts.ReleaseType, args.ProImageOpts.GrafanaVersion),
		})

	if args.ProImageOpts.Push {
		panic("TODO: push to container registry")
	}
	stdout, err := container.Stdout(context.Background())
	fmt.Printf("\n\naaaaaaa stdout=%+v err=%+v\n\n", stdout, err)

	stderr, err := container.Stderr(context.Background())
	fmt.Printf("\n\naaaaaaa stderr=%+v err=%+v\n\n", stderr, err)

	return nil
}
