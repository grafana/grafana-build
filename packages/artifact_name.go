package packages

import (
	"context"

	"github.com/grafana/grafana-build/arguments"
	"github.com/grafana/grafana-build/executil"
	"github.com/grafana/grafana-build/flags"
	"github.com/grafana/grafana-build/pipeline"
)

func ArtifactFilename(ctx context.Context, a pipeline.Artifact, h pipeline.StateHandler, ext string) (string, error) {
	name, err := a.Option(flags.PackageName)
	if err != nil {
		return "", err
	}

	distro, err := a.Option(flags.PackageDistribution)
	if err != nil {
		return "", err
	}

	buildID, err := h.String(ctx, arguments.BuildID)
	if err != nil {
		return "", err
	}

	version, err := h.String(ctx, arguments.Version)
	if err != nil {
		return "", err
	}

	opts := NameOpts{
		// Name is the name of the product in the package. 99% of the time, this will be "grafana" or "grafana-enterprise".
		Name:      name,
		Version:   version,
		BuildID:   buildID,
		Distro:    executil.Distribution(distro),
		Extension: ext,
	}

	return FileName(opts)
}
