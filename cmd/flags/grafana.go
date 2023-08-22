package flags

import "github.com/urfave/cli/v2"

// Grafana flags are flags that are required when working with the grafana source code.
var Grafana = []cli.Flag{
	&cli.BoolFlag{
		Name:     "grafana",
		Usage:    "If set, initialize Grafana",
		Required: false,
		Value:    true,
	},
	&cli.StringFlag{
		Name:     "grafana-dir",
		Usage:    "Local Grafana dir to use, instead of git clone",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "grafana-repo",
		Usage:    "Grafana repo to clone, not valid if --grafana-dir is set",
		Required: false,
		Value:    "https://github.com/grafana/grafana.git",
	},
	&cli.StringFlag{
		Name:     "grafana-ref",
		Usage:    "Grafana ref to clone, not valid if --grafana-dir is set",
		Required: false,
		Value:    "main",
	},
	&cli.BoolFlag{
		Name:  "enterprise",
		Usage: "If set, initialize Grafana Enterprise",
		Value: false,
	},
	&cli.StringFlag{
		Name:     "enterprise-dir",
		Usage:    "Local Grafana Enterprise dir to use, instead of git clone",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "enterprise-repo",
		Usage:    "Grafana Enterprise repo to clone, not valid if --grafana-dir is set",
		Required: false,
		Value:    "https://github.com/grafana/grafana-enterprise.git",
	},
	&cli.StringFlag{
		Name:     "enterprise-ref",
		Usage:    "Grafana Enterprise ref to clone, not valid if --enterprise-dir is set",
		Required: false,
		Value:    "main",
	},
	&cli.StringFlag{
		Name:     "build-id",
		Usage:    "Build ID to use, by default will be what is defined in package.json",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "github-token",
		Usage:    "Github token to use for git cloning, by default will be pulled from GitHub",
		Required: false,
	},
	&cli.StringFlag{
		Name:  "version",
		Usage: "Explicit version number. If this is not set then one with will auto-detected based on the source repository",
	},
	&cli.StringSliceFlag{
		Name:    "env",
		Aliases: []string{"e"},
		Usage:   "Set a build-time environment variable using the same syntax as 'docker run'. Example: `--env=GOOS=linux --env=GOARCH=amd64`",
	},
	&cli.StringSliceFlag{
		Name:  "go-tags",
		Usage: "Sets the go `-tags` flag when compiling the backend",
	},
	&cli.StringFlag{
		Name:  "yarn-cache",
		Usage: "If there is a yarn cache directory, then mount that when running 'yarn install' instead of creating a cache directory",
	},
}
