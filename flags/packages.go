package flags

import (
	"github.com/grafana/grafana-build/packages"
	"github.com/grafana/grafana-build/pipeline"
)

const (
	PackageName   pipeline.FlagOption = "package-name"
	Distribution  pipeline.FlagOption = "distribution"
	Static        pipeline.FlagOption = "static"
	Enterprise    pipeline.FlagOption = "enterprise"
	WireTag       pipeline.FlagOption = "wire-tag"
	GoTags        pipeline.FlagOption = "go-tag"
	GoExperiments pipeline.FlagOption = "go-experiments"
	Sign          pipeline.FlagOption = "sign"

	// Pretty much only used to set the deb or RPM internal package name (and file name) to `{}-nightly` and/or `{}-rpi`
	Nightly pipeline.FlagOption = "nightly"
	RPI     pipeline.FlagOption = "rpi"
)

// These are the flags that packages (targz, deb, rpm, docker) must have.
// Essentially they must have all of the same things that the targz package have.
var PackageNameFlags = []pipeline.Flag{
	{
		Name: "grafana",
		Options: map[pipeline.FlagOption]any{
			DockerRepositories: []string{"grafana-image-tags", "grafana-oss-image-tags"},
			PackageName:        string(packages.PackageGrafana),
			Enterprise:         false,
			WireTag:            "oss",
			GoExperiments:      []string{},
			GoTags:             []string{},
		},
	},
	{
		Name: "enterprise",
		Options: map[pipeline.FlagOption]any{
			DockerRepositories: []string{"grafana-enterprise-image-tags"},
			PackageName:        string(packages.PackageEnterprise),
			Enterprise:         true,
			WireTag:            "enterprise",
			GoExperiments:      []string{},
			GoTags:             []string{},
		},
	},
	{
		Name: "pro",
		Options: map[pipeline.FlagOption]any{
			DockerRepositories: []string{"grafana-pro-image-tags"},
			PackageName:        string(packages.PackagePro),
			Enterprise:         true,
			WireTag:            "pro",
			GoExperiments:      []string{},
			GoTags:             []string{"pro"},
		},
	},
	{
		Name: "boring",
		Options: map[pipeline.FlagOption]any{
			DockerRepositories: []string{"grafana-enterprise-image-tags"},
			PackageName:        string(packages.PackageEnterpriseBoring),
			Enterprise:         true,
			WireTag:            "enterprise",
			GoExperiments:      []string{"boringcrypto"},
		},
	},
}

var SignFlag = pipeline.Flag{
	Name: "sign",
	Options: map[pipeline.FlagOption]any{
		Sign: true,
	},
}

var NightlyFlag = pipeline.Flag{
	Name: "nightly",
	Options: map[pipeline.FlagOption]any{
		Nightly: true,
	},
}

func StdPackageFlags() []pipeline.Flag {
	distros := DistroFlags()
	names := PackageNameFlags

	return JoinFlags(
		distros,
		names,
	)
}
