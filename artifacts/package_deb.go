package artifacts

import (
	"github.com/grafana/grafana-build/arguments"
	"github.com/grafana/grafana-build/flags"
	"github.com/grafana/grafana-build/pipeline"
)

var Deb = pipeline.Artifact{
	Name:     "deb",
	Requires: []pipeline.Artifact{Tarball},
	Flags:    flags.StdPackageFlags(),
	Arguments: []pipeline.Argument{
		arguments.BuildID,
		arguments.Version,
	},
	FileNameFunc: PackageWithExtension("deb"),
}
