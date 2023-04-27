package main

import (
	"github.com/urfave/cli/v2"
)

var FlagPackage = &cli.StringSliceFlag{
	Name:  "package",
	Usage: "Path to a grafana.tar.gz package used as input. This command will process each package provided separately and produce an equal number of applicable outputs.",
}

// PackageInputFlags are used for commands that require a grafana package as input.
// These commands are exclusively used outside of the CI process and are typically used in the CD process where a grafana.tar.gz has already been created.
var PackageInputFlags = []cli.Flag{
	FlagPackage,
}

// PublishFlags are flags that are used in commands that create artifacts.
// Anything that creates an artifact should have the option to specify a local folder destination or a remote destination.
var PublishFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "destination",
		Usage:   "full URL to upload the artifacts to (examples: '/tmp/package.tar.gz', 'file://package.tar.gz', 'file:///tmp/package.tar.gz', 'gs://bucket/grafana/')",
		Aliases: []string{"d"},
		Value:   "file://dist",
	},
	&cli.StringFlag{
		Name:     "gcp-service-account-key-base64",
		Usage:    "Provides a service-account key encoded in base64 to use to authenticate with the Google Cloud SDK",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "gcp-service-account-key",
		Usage:    "Provides a service-account keyfile to use to authenticate with the Google Cloud SDK. If not provided or is empty, then $XDG_CONFIG_HOME/gcloud will be mounted in the container",
		Required: false,
	},
}

// GrafanaFlags are flags that are required when working with the grafana source code.
var GrafanaFlags = []cli.Flag{
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
}

var FlagDistros = &cli.StringSliceFlag{
	Name:  "distro",
	Usage: "See the list of distributions with 'go tool dist list'. For variations of the same distribution, like 'armv6' or 'armv7', append an extra path part. Example: 'linux/arm/v6', or 'linux/amd64/v3'.",
	Value: cli.NewStringSlice("linux/amd64", "linux/arm64"),
}

// PackageFlags are flags that are used when building packages or similar artifacts (like binaries) for different distributions
// from the grafana source code.
var PackageFlags = []cli.Flag{
	FlagDistros,
}

var DefaultFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:    "verbose",
		Aliases: []string{"v"},
		Usage:   "Increase log verbosity. WARNING: This setting could potentially log sensitive data.",
		Value:   false,
	},
}

// JoinFlags combines several slices of flags into one slice of flags.
func JoinFlags(f ...[]cli.Flag) []cli.Flag {
	flags := []cli.Flag{}
	for _, v := range f {
		flags = append(flags, v...)
	}

	return flags
}

func JoinFlagsWithDefault(f ...[]cli.Flag) []cli.Flag {
	// Kind of gross but ensures that DeafultFlags are registered before any others.
	return JoinFlags(append([][]cli.Flag{DefaultFlags}, f...)...)
}
