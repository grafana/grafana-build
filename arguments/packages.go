package arguments

import (
	"context"

	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
)

var flagBuildID = &cli.StringFlag{
	Name:  "build-id",
	Usage: "Build ID to use in package names. If not set, then a random string is generated",
}

var BuildID = pipeline.Argument{
	Name:        "build-id",
	Description: "The grafana backend binaries ('grafana', 'grafana-cli', 'grafana-server') in a directory",
	Flags: []cli.Flag{
		flagBuildID,
	},
	ValueFunc: func(ctx context.Context, opts *pipeline.ArgumentOpts) (any, error) {
		return opts.CLIContext.String("build-id"), nil
	},
}

var flagVersion = &cli.StringFlag{
	Name:  "version",
	Usage: "Explicit version number. If this is not set then one with will auto-detected based on the source repository",
}

var Version = pipeline.Argument{
	Name:        "version",
	Description: "The version string that is shown in the UI, in the CLI, and in package metadata",
	Flags: []cli.Flag{
		flagVersion,
	},
	Requires: []pipeline.Argument{
		GrafanaDirectory,
	},
	ValueFunc: func(ctx context.Context, opts *pipeline.ArgumentOpts) (any, error) {
		v := opts.CLIContext.String("version")
		if v != "" {
			return v, nil
		}
		src, err := opts.State.Directory(ctx, GrafanaDirectory)
		if err != nil {
			return "", err
		}

		return containers.GetJSONValue(ctx, opts.Client, src, "package.json", "version")
	},
}
