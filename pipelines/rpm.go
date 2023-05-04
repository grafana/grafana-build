package pipelines

import (
	"context"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

// RPM uses the grafana package given by the '--package' argument and creates a .rpm installer.
// It accepts publish args, so you can place the file in a local or remote destination.
func RPM(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	return PackageInstaller(ctx, d, args, InstallerOpts{
		PackageType: "rpm",
		ConfigFiles: [][]string{
			{"/src/packaging/rpm/sysconfig/grafana-server", "/pkg/etc/sysconfig/grafana-server"},
			{"/src/packaging/rpm/init.d/grafana-server", "/pkg/etc/init.d/grafana-server"},
			{"/src/packaging/rpm/systemd/grafana-server.service", "/pkg/usr/lib/systemd/system/grafana-server.service"},
		},
		AfterInstall: "/src/packaging/rpm/control/postinst",
		Depends: []string{
			"/sbin/service",
			"chkconfig",
			"fontconfig",
			"freetype",
			"urw-fonts",
		},
		ExtraArgs: []string{
			"--rpm-posttrans=/src/packaging/rpm/control/posttrans",
			"--rpm-digest=sha256",
		},
		EnvFolder: "/pkg/etc/sysconfig",
		RPMSign:   args.GPGOpts.Sign,
		Container: containers.RPMContainer(d, args.GPGOpts),
	})
}
