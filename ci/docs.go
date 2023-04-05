package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"github.com/urfave/cli/v2"
)

const localTechdocsImage = "squidfunk/mkdocs-material:9.1.4"
const techdocsImage = "europe-west4-docker.pkg.dev/grafana-backstage-375210/grafana-backstage/techdocs:latest"
const backstageBucketName = "backstage-techdocs-storage"
const backstageComponentName = "grafana-build"

// Serve starts a local webserver and watch for updates
func serveDocsAction(cliCtx *cli.Context) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	//nolint:gosec
	cmd := exec.CommandContext(cliCtx.Context, "docker", "run",
		"--platform", "linux/amd64",
		"-w", "/src",
		"-v", pwd+":/src",
		"-p", "8000:8000",
		"--rm",
		localTechdocsImage,
		"serve", "--dev-addr", "0.0.0.0:8000")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func buildTechDocs(ctx context.Context, dc *dagger.Client) error {
	googleCredentialsPath := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if googleCredentialsPath == "" {
		return fmt.Errorf("no GOOGLE_APPLICATION_CREDENTIALS set")
	}
	src := dc.Host().Directory(".")
	techdocContainer := dc.Container(containerOpts).From(techdocsImage).
		WithWorkdir("/src").
		WithMountedDirectory("/src", src)
	techdocContainer = techdocContainer.WithExec([]string{"build"})
	if _, err := techdocContainer.ExitCode(ctx); err != nil {
		return err
	}

	targetPathPrefix := fmt.Sprintf("default/component/%s", backstageComponentName)
	googleCredentialsBase := filepath.Base(googleCredentialsPath)

	gcsContainer := dc.Container(containerOpts).From("google/cloud-sdk:alpine").
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithDirectory("/data", techdocContainer.Directory("/src/site")).
		WithEnvVariable("GOOGLE_APPLICATION_CREDENTIALS", "./"+googleCredentialsBase).
		WithExec([]string{"gcloud", "auth", "login", "--cred-file", "./" + googleCredentialsBase}).
		WithExec([]string{"gcloud", "auth", "list"}).
		WithExec([]string{"gsutil", "-m", "rsync", "-d", "-r", "/data", fmt.Sprintf("gs://%s/%s", backstageBucketName, targetPathPrefix)})

	if _, err := gcsContainer.ExitCode(ctx); err != nil {
		return fmt.Errorf("failed to upload techdocs: %w", err)
	}
	return nil
}
