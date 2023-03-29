package main

import (
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

const techdocsImage = "squidfunk/mkdocs-material:9.1.4"

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
		techdocsImage,
		"serve", "--dev-addr", "0.0.0.0:8000")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
