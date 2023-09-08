package artifacts

import (
	"github.com/grafana/grafana-build/backend"
	"github.com/grafana/grafana-build/pipeline"
)

var Backend = &pipeline.Artifact{
	Name: "backend",
	Arguments: []pipeline.Argument{
		ArgumentGrafanaDirectory,
		ArgumentEnterpriseDirectory,
		ArgumentGoImage,
	},
	Builder:   backend.Builder,
	BuildFunc: backend.Build,

	Publisher:   nil,
	PublishFunc: nil,
}
