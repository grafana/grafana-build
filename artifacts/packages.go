package artifacts

import (
	"context"

	"github.com/grafana/grafana-build/arguments"
	"github.com/grafana/grafana-build/backend"
	"github.com/grafana/grafana-build/flags"
	"github.com/grafana/grafana-build/packages"
	"github.com/grafana/grafana-build/pipeline"
)

type PackageDetails struct {
	Name         packages.Name
	Enterprise   bool
	Version      string
	BuildID      string
	Distribution backend.Distribution
}

func GetPackageDetails(ctx context.Context, options *pipeline.OptionsHandler, state pipeline.StateHandler) (PackageDetails, error) {
	distro, err := options.String(flags.Distribution)
	if err != nil {
		return PackageDetails{}, err
	}
	version, err := state.String(ctx, arguments.Version)
	if err != nil {
		return PackageDetails{}, err
	}
	buildID, err := state.String(ctx, arguments.BuildID)
	if err != nil {
		return PackageDetails{}, err
	}

	name, err := options.String(flags.PackageName)
	if err != nil {
		return PackageDetails{}, err
	}

	enterprise, err := options.Bool(flags.Enterprise)
	if err != nil {
		return PackageDetails{}, err
	}

	return PackageDetails{
		Name:         packages.Name(name),
		Version:      version,
		BuildID:      buildID,
		Distribution: backend.Distribution(distro),
		Enterprise:   enterprise,
	}, nil
}
