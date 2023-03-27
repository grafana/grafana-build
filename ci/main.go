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

func mainAction(c *cli.Context) error {
	ctx := c.Context
	dc, err := dagger.Connect(ctx, dagger.WithLogOutput(log.Default().Writer()))
	if err != nil {
		return err
	}
	defer dc.Close()

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
	app.RunContext(ctx, os.Args)
}
