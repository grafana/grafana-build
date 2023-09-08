package artifacts

import "github.com/grafana/grafana-build/pipeline"

type Registerer interface {
	Register(*pipeline.Artifact) error
}
