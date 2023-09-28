package artifacts

import (
	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
)

// An ArtifactFlag should provide all of the necessary arguments to produce an artifact
// dleimited by colons.
// Examples:
// * targz:linux/amd64 -- Will produce a "Grafana" tar.gz for "linux/amd64".
// * targz:enterprise:linux/amd64 -- Will produce a "Grafana" tar.gz for "linux/amd64".
type ArtifactFlag struct {
	cli.StringSliceFlag

	Artifacts map[string]pipeline.Argument
}
