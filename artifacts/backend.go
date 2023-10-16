package artifacts

import (
	"github.com/grafana/grafana-build/arguments"
	"github.com/grafana/grafana-build/pipeline"
)

var Backend = pipeline.Artifact{
	Name: "backend",
	Arguments: []pipeline.Argument{
		arguments.GrafanaDirectory,
		arguments.EnterpriseDirectory,
		arguments.GoImage,
	},
	// Builder:   backend.Builder,
	// BuildFunc: backend.Build,

	// Publisher:   nil,
	// PublishFunc: nil,
}
