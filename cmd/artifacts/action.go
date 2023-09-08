package artifacts

import (
	"github.com/grafana/grafana-build/pipeline"
	"github.com/urfave/cli/v2"
)

// Register adds the pipeline artifacts to the urfave/cli app.
func Register(r Registerer, a ...*pipeline.Artifact) error {
	for _, v := range a {
		if err := r.Register(v); err != nil {
			return err
		}
	}

	return nil
}

// Action returns the urfave CLI action that handles building an artifact
func BuildAction(a *pipeline.Artifact) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		return nil
	}
}

func Command(artifacts []*pipeline.Artifact) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		return nil
	}
}

func ArtifactFlags(artifacts []*pipeline.Artifact) []cli.Flag {
	return nil
}
