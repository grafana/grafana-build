package flags

import "github.com/urfave/cli/v2"

var BuildID = &cli.StringFlag{
	Name:  "build-id",
	Usage: "The build ID. Used to correlate builds with CI.",
	Value: "",
}
