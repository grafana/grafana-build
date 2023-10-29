package exe

import (
	"fmt"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

func Build(d *dagger.Client, builder *dagger.Container, targz *dagger.File, enterprise bool) *dagger.File {
	f := containers.ExtractedArchive(d, targz)

	var (
		app     = "GrafanaOSS"
		slug    = "grafana"
		license = "AGPLv3.rtf" // TODO: we should probably prefer that the installer shows the license.txt in the repo instead.
	)

	if enterprise {
		app = "GrafanaEnterprise"
		slug = "grafana-enterprise"
		license = "grafana-enterprise.rtf"
	}

	nsisArgs := []string{
		"makensis",
		"-V3",
		fmt.Sprintf("-DAPPNAME=%s", app),
		fmt.Sprintf("-DAPPNAME_SLUG=%s", slug),
		fmt.Sprintf("-DLICENSE=%s", license),
		"grafana.nsis",
	}

	return builder.
		WithMountedDirectory("/src/grafana", f).
		WithWorkdir("/src").
		WithExec([]string{"ls", "-al"}).
		WithExec([]string{"ls", "-al", "winimg"}).
		WithExec(nsisArgs).
		File("/src/grafana-setup.exe")
}
