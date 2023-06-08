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
		Name:  "gcp-service-account-key-base64",
		Usage: "Provides a service-account key encoded in base64 to use to authenticate with the Google Cloud SDK",
	},
	&cli.StringFlag{
		Name:  "gcp-service-account-key",
		Usage: "Provides a service-account keyfile to use to authenticate with the Google Cloud SDK. If not provided or is empty, then $XDG_CONFIG_HOME/gcloud will be mounted in the container",
	},
	&cli.BoolFlag{
		Name:  "checksum",
		Usage: "When enabled, also creates a `.sha256' checksum file in the destination that matches the checksum of the artifact(s) produced",
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
	&cli.StringSliceFlag{
		Name:    "env",
		Aliases: []string{"e"},
		Usage:   "Set a build-time environment variable using the same syntax as 'docker run'. Example: `--env=GOOS=linux --env=GOARCH=amd64`",
	},
	&cli.StringSliceFlag{
		Name:  "go-tags",
		Usage: "Sets the go `-tags` flag when compiling the backend",
	},
}

// DockerFlags are used when producing docker images.
var DockerFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "registry",
		Usage: "Prefix the image name with the registry provided",
		Value: "docker.io",
	},
	&cli.StringFlag{
		Name:  "alpine-base",
		Usage: "The alpine image to use as the base image when building the Alpine version of the Grafana docker image",
		Value: "alpine:latest",
	},
	&cli.StringFlag{
		Name:  "ubuntu-base",
		Usage: "The Ubuntu image to use as the base image when building the Ubuntu version of the Grafana docker image",
		Value: "ubuntu:latest",
	},
	&cli.StringFlag{
		Name:  "alpine-base-armv7",
		Usage: "The alpine image to use as the base image when building the Alpine (armv7) version of the Grafana docker image",
		Value: "arm32v7/alpine:latest",
	},
	&cli.StringFlag{
		Name:  "ubuntu-base-armv7",
		Usage: "The Ubuntu image to use as the base image when building the Ubuntu (armv7) version of the Grafana docker image",
		Value: "arm32v7/ubuntu:latest",
	},
	&cli.StringFlag{
		Name:  "alpine-base-arm64",
		Usage: "The alpine image to use as the base image when building the Alpine (arm64) version of the Grafana docker image",
		Value: "arm64v8/alpine:latest",
	},
	&cli.StringFlag{
		Name:  "ubuntu-base-arm64",
		Usage: "The Ubuntu image to use as the base image when building the Ubuntu (arm64) version of the Grafana docker image",
		Value: "arm64v8/ubuntu:latest",
	},
}

var FlagDistros = &cli.StringSliceFlag{
	Name:  "distro",
	Usage: "See the list of distributions with 'go tool dist list'. For variations of the same distribution, like 'armv6' or 'armv7', append an extra path part. Example: 'linux/arm/v6', or 'linux/amd64/v3'.",
	Value: cli.NewStringSlice(DefaultDistros...),
}

var GPGFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "gpg-private-key-base64",
		Usage: "Provides a private key encoded in base64 to use to for GPG signing",
	},
	&cli.StringFlag{
		Name:  "gpg-public-key-base64",
		Usage: "Provides a public key encoded in base64 to use to for GPG signing",
	},
	&cli.StringFlag{
		Name:  "gpg-passphrase-base64",
		Usage: "Provides a private key passphrase encoded in base64 to use to for GPG signing",
	},
	&cli.BoolFlag{
		Name:  "sign",
		Usage: "Enable GPG signing of RPM packages",
	},
}

// PackageFlags are flags that are used when building packages or similar artifacts (like binaries) for different distributions
// from the grafana source code.
var PackageFlags = []cli.Flag{
	FlagDistros,
	&cli.StringFlag{
		Name:  "edition",
		Usage: "Simply alters the naming of the '.tar.gz' package. The string set will override the '-{flavor}' part of the package name",
	},
}

var ProImageFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "github-token",
		Usage:    "Github token to use for git cloning, by default will be pulled from GitHub",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "deb",
		Usage:    "The Grafana debian package that will be used to build the pro image",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "grafana-version",
		Usage:    "The Grafana version",
		Required: true,
	},
	&cli.StringFlag{
		Name:  "release-type",
		Usage: "The Grafana release type",
		Value: "prerelease",
	},
	&cli.BoolFlag{
		Name:  "push",
		Usage: "Push the built image to the container registry",
		Value: false,
	},
}

var DefaultFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "platform",
		Usage: "The buildkit / dagger platform to run containers when building the backend",
		Value: DefaultPlatform,
	},
	&cli.BoolFlag{
		Name:    "verbose",
		Aliases: []string{"v"},
		Usage:   "Increase log verbosity. WARNING: This setting could potentially log sensitive data",
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
