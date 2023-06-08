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
		WithDirectory("/src", hostedGrafanaRepo).
		WithFile("/src/grafana.deb", debianPackageFile, dagger.ContainerWithFileOpts{}).
		WithWorkdir("/src").
		WithExec([]string{
			"/bin/sh", "-c",
			"docker build --platform=linux/amd64 . -f ./cmd/hgrun/Dockerfile -t hgrun:latest",
		}).
		WithExec([]string{
			"/bin/sh", "-c",
			fmt.Sprintf("docker build --platform=linux/amd64 --build-arg=RELEASE_TYPE=%s --build-arg=GRAFANA_VERSION=%s --build-arg=HGRUN_IMAGE=hgrun:latest . -f ./docker/hosted-grafana-all/Dockerfile -t hostedgrafana:latest",
				args.ProImageOpts.ReleaseType,
				args.ProImageOpts.GrafanaVersion,
			),
		})

	if args.ProImageOpts.Push {
		panic("TODO: push to container registry")
	}

	exitCode, err := container.ExitCode(ctx)
	if err != nil {
		return fmt.Errorf("getting container exit code: %w", err)
	}

	if exitCode != 0 {
		stderr, err := container.Stderr(ctx)
		if err != nil {
			return fmt.Errorf("getting container stderr: exitCode=%d %w", exitCode, err)
		}
		stdout, err := container.Stdout(ctx)
		if err != nil {
			return fmt.Errorf("getting container stdout: exitCode=%d %w", exitCode, err)
		}
		return fmt.Errorf("container exit code is not 0: exitCode=%d stderr=%s stdout=%s", exitCode, stderr, stdout)
	}

	return nil
}
