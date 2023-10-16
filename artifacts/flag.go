package artifacts

import (
	"strings"

	"log/slog"

	"github.com/grafana/grafana-build/pipeline"
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

	flags := []cli.Flag{
		artifactsFlag,
		buildFlag,
		publishFlag,
		platformFlag,
	}

	// All of these artifacts are the registered artifacts. These should mostly stay the same no matter what.
	artifacts := r.Artifacts()

	// Add all of the CLI flags that are defined by each artifact's arguments.
	m := map[string]cli.Flag{}
	// For artifact arguments that specify flags, we'll coalesce them here and add them to the list of flags.
	for _, artifact := range artifacts {
		for _, arg := range artifact.Arguments {
			for _, f := range arg.Flags {
				fn := strings.Join(f.Names(), ",")
				m[fn] = f
				slog.Debug("global flag added by argument in artifact", "flag", fn, "arg", arg.Name, "artifact", artifact.Name)
			}
		}
	}

	for _, v := range m {
		flags = append(flags, v)
	}

	return flags
}

// The ArtifactsFromStrings function should provide all of the necessary arguments to produce each artifact
// dleimited by colons. It's a repeated flag, so all permutations are stored in 1 instance of the ArtifactsFlag struct.
// Examples:
// * targz:linux/amd64 -- Will produce a "Grafana" tar.gz for "linux/amd64".
// * targz:enterprise:linux/amd64 -- Will produce a "Grafana" tar.gz for "linux/amd64".
func ArtifactsFromStrings(a []string, registered []pipeline.Artifact) ([]pipeline.Artifact, error) {
	artifacts := make([]pipeline.Artifact, len(a))
	for i, v := range a {
		n, err := Parse(v, registered)
		if err != nil {
			return nil, err
		}

		artifacts[i] = n
	}

	return artifacts, nil
}
