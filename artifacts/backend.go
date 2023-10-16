package artifacts

import (
	"github.com/grafana/grafana-build/arguments"
	"github.com/grafana/grafana-build/backend"
	"github.com/grafana/grafana-build/pipeline"
)

var Backend = pipeline.Artifact{
	Name: "backend",
	Type: pipeline.ArtifactTypeDirectory,
	Arguments: []pipeline.Argument{
		arguments.GrafanaDirectory,
		arguments.EnterpriseDirectory,
		arguments.GoImage,
	},
	Builder:      backend.Builder,
	BuildDirFunc: backend.Build,

	// Publisher:   nil,
	// PublishFunc: nil,
}
