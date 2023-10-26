package artifacts

import (
	"strings"

	"log/slog"

	"github.com/grafana/grafana-build/cmd/flags"
	"github.com/urfave/cli/v2"
)

func ArtifactFlags(r Registerer) []cli.Flag {
	artifactsFlag := &cli.StringSliceFlag{
		Name:    "artifacts",
		Aliases: []string{"a"},
	}

	buildFlag := &cli.BoolFlag{
		Name:  "build",
		Value: true,
	}

	publishFlag := &cli.BoolFlag{
		Name:  "publish",
		Usage: "If true, then the artifacts that are built will be published. If `--build=false` and the artifacts are found in the --destination, then those artifacts are not built and are published instead.",
		Value: true,
	}

	platformFlag := &cli.StringFlag{
		Name:  "platform",
		Value: "linux/amd64",
	}

	flags := flags.Join(
		[]cli.Flag{
			artifactsFlag,
			buildFlag,
			publishFlag,
			platformFlag,
		},
		flags.PublishFlags,
		flags.ConcurrencyFlags,
	)

	// All of these artifacts are the registered artifacts. These should mostly stay the same no matter what.
	initializers := r.Initializers()

	// Add all of the CLI flags that are defined by each artifact's arguments.
	m := map[string]cli.Flag{}

	// For artifact arguments that specify flags, we'll coalesce them here and add them to the list of flags.
	for _, n := range initializers {
		for _, arg := range n.Arguments {
			for _, f := range arg.Flags {
				fn := strings.Join(f.Names(), ",")
				m[fn] = f
				slog.Debug("global flag added by argument in artifact", "flag", fn, "arg", arg.Name)
			}
		}
	}

	for _, v := range m {
		flags = append(flags, v)
	}

	return flags
}
