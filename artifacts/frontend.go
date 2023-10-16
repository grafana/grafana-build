package artifacts

import (
	"github.com/grafana/grafana-build/backend"
	"github.com/grafana/grafana-build/pipeline"
)

var Frontend = pipeline.Artifact{
	Name:         "frontend",
	Type:         pipeline.ArtifactTypeDirectory,
	Builder:      backend.Builder,
	BuildDirFunc: backend.Build,
}
