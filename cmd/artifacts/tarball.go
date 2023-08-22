package artifacts

import "github.com/grafana/grafana-build/pipeline"

var Tarball = &pipeline.Artifact{
	Name:     "tarball",
	Requires: []*pipeline.Artifact{Backend, Frontend},
}
