package main

import "github.com/urfave/cli/v2"

var FlagDistro = &cli.StringFlag{
	Name:  "distro",
	Usage: "See the list of distributions with 'go tool dist list'",
	Value: "linux/amd64",
}
