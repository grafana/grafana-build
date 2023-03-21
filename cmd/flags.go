package main

import (
	"github.com/urfave/cli/v2"
)

var FlagDistros = &cli.StringSliceFlag{
	Name:  "distro",
	Usage: "See the list of distributions with 'go tool dist list'. For variations of the same distribution, like 'armv6' or 'armv7', append an extra path part. Example: 'linux/arm/v6', or 'linux/amd64/v3'.",
	Value: cli.NewStringSlice(DefaultDistros...),
}
