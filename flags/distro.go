package flags

import (
	"github.com/grafana/grafana-build/backend"
	"github.com/grafana/grafana-build/pipeline"
)

const (
	FlagDistribution = "distro"
)

var StaticDistributions = []backend.Distribution{
	backend.DistLinuxAMD64,
	backend.DistLinuxARM64,
	backend.DistLinuxARMv7,
	backend.DistLinuxRISCV64,
}

var DynamicDistributions = []backend.Distribution{
	backend.DistDarwinAMD64,
	backend.DistDarwinARM64,
	backend.DistWindowsAMD64,
	backend.DistWindowsARM64,
	backend.DistLinuxAMD64Dynamic,
	backend.DistLinuxAMD64DynamicMusl,
}

func DistroFlags() []pipeline.Flag {
	// These distributions have specific options that set some stuff.
	f := []pipeline.Flag{
		{
			Name: string(backend.DistLinuxARMv6),
			Options: map[pipeline.FlagOption]any{
				Distribution: string(backend.DistLinuxARMv6),
				Static:       true,
				RPI:          true,
			},
		},
	}

	for _, v := range StaticDistributions {
		d := string(v)
		f = append(f, pipeline.Flag{
			Name: d,
			Options: map[pipeline.FlagOption]any{
				Distribution: d,
				Static:       true,
			},
		})
	}
	for _, v := range DynamicDistributions {
		d := string(v)
		f = append(f, pipeline.Flag{
			Name: d,
			Options: map[pipeline.FlagOption]any{
				Distribution: d,
				Static:       false,
			},
		})
	}

	return f
}
