package pipelines

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

func WithRPMSignature(ctx context.Context, d *dagger.Client, opts *containers.GPGOpts, installers map[string]*dagger.File) (map[string]*dagger.File, error) {
	out := make(map[string]*dagger.File, len(installers))

	for dst, file := range installers {
		if filepath.Ext(dst) != ".rpm" {
			log.Println(dst, "is not an rpm, it is", filepath.Ext(dst))
			out[dst] = file
			continue
		}

		container, err := containers.GPGContainer(d, opts)
		if err != nil {
			return nil, err
		}
		f := container.
			WithEnvVariable("dst", dst).
			WithMountedFile("/src/package.rpm", file).
			WithExec([]string{"rpm", "--addsign", "/src/package.rpm"}).
			WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("rpm --checksig %s | grep -qE 'digests signatures OK|pgp.+OK'", "/src/package.rpm")}).
			File("/src/package.rpm")

		out[dst] = f
	}

	return out, nil
}

// RPM uses the grafana package given by the '--package' argument and creates a .rpm installer.
// It accepts publish args, so you can place the file in a local or remote destination.
func RPM(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	installers, err := PackageInstaller(ctx, d, args, InstallerOpts{
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
		Container: containers.RPMContainer(d),
	})

	if err != nil {
		return err
	}

	if args.GPGOpts.Sign {
		f, err := WithRPMSignature(ctx, d, args.GPGOpts, installers)
		if err != nil {
			return err
		}
		installers = f
	}

	return PublishInstallers(ctx, d, args, installers)
}
