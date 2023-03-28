package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"dagger.io/dagger"
	"github.com/urfave/cli/v2"
)

func init() {
	log.SetOutput(os.Stderr)
}

func lintProject(ctx context.Context, dc *dagger.Client) error {
	workDir := dc.Host().Directory(".")
	container := dc.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).
		From("golangci/golangci-lint:v1.52.2").
		WithWorkdir("/src").
		WithMountedDirectory("/src", workDir).
		WithExec([]string{"golangci-lint", "run"})

	if _, err := container.ExitCode(ctx); err != nil {
		return err
	}
	return nil
}

func mainAction(cctx *cli.Context) (rerr error) {
	ctx := cctx.Context
	dc, err := dagger.Connect(ctx, dagger.WithLogOutput(log.Default().Writer()))
	if err != nil {
		return err
	}
	defer func() {
		if err := dc.Close(); rerr == nil && err != nil {
			rerr = err
		}
	}()

	if err := lintProject(ctx, dc); err != nil {
		return fmt.Errorf("linting failed: %w", err)
	}

	workDir := dc.Host().Directory(".")
	goContainer := dc.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).
		From("golang:1.20.2")

	// If the gopath is set, mount the pkg folder from there so that we can re-use as much as possible from the system's caching:
	rawGoPath, err := exec.CommandContext(ctx, "go", "env", "GOPATH").CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to determine GOPATH: %w", err)
	}
	goPath := strings.TrimSpace(string(rawGoPath))
	if goPath != "" {
		log.Print("Mounting GOPATH/pkg from host system")
		hostGoPath := dc.Host().Directory(goPath)
		goContainer = goContainer.WithMountedDirectory("/go/pkg", hostGoPath.Directory("pkg"))
	} else {
		log.Print("Mounting cache volume as /go/pkg")
		goCache := dc.CacheVolume("ci-go")
		goContainer = goContainer.WithMountedCache("/go/pkg", goCache)
	}

	log.Print("Executing tests")
	if _, err := goContainer.
		WithMountedDirectory("/src", workDir).
		WithWorkdir("/src").
		WithExec([]string{"go", "test", "./...", "-v"}).
		ExitCode(ctx); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}
	return nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	app := cli.App{
		DefaultCommand: "main",
		Commands: []*cli.Command{
			{
				Name:   "main",
				Action: mainAction,
			},
		},
	}
	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err.Error())
	}
}
