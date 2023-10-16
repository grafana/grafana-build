package flags

import (
	"github.com/grafana/grafana-build/executil"
	"github.com/grafana/grafana-build/pipeline"
)

const (
	PackageName         = "package-name"
	PackageDistribution = "package-distribution"
	PackageBuildID      = "package-build-id"
	PackageEnterprise   = "package-enterprise"
)

// These are the flags that packages (targz, deb, rpm, docker) must have.
// Essentially they must have all of the same things that the targz package have.
var PackageNameFlags = []pipeline.Flag{
	{
		Name: "grafana",
		Options: map[string]string{
			PackageName: "grafana",
		},
	},
	{
		Name: "enterprise",
		Options: map[string]string{
			PackageName: "grafana-enterprise",
		},
	},
	{
		Name: "pro",
		Options: map[string]string{
			PackageName: "grafana-pro",
		},
	},
	{
		Name: "boring",
		Options: map[string]string{
			PackageName: "grafana-enterprise-boringcrypto",
		},
	},
}

var Distributions = []executil.Distribution{
	executil.DistLinuxAMD64,
	executil.DistLinuxARM64,
	executil.DistLinuxARMv6,
	executil.DistLinuxARMv7,
	executil.DistDarwinAMD64,
	executil.DistDarwinARM64,
	executil.DistWindowsAMD64,
	executil.DistWindowsARM64,
	executil.DistLinuxRISCV64,
}

func DistroFlags() []pipeline.Flag {
	f := make([]pipeline.Flag, len(Distributions))
	for i, v := range Distributions {
		d := string(v)
		f[i] = pipeline.Flag{
			Name: d,
			Options: map[string]string{
				PackageDistribution: d,
			},
		}
	}

	return f
}

func JoinFlags(f ...[]pipeline.Flag) []pipeline.Flag {
	r := []pipeline.Flag{}
	for _, v := range f {
		r = append(r, v...)
	}

	return r
}

func StdPackageFlags() []pipeline.Flag {
	distros := DistroFlags()
	names := PackageNameFlags

	return JoinFlags(
		distros,
		names,
	)
}
