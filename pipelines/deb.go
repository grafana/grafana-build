package pipelines

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
)

// Deb uses the grafana package given by the '--package' argument and creates a .deb installer.
// It accepts publish args, so you can place the file in a local or remote destination.
func Deb(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts)
	if err != nil {
		return err
	}

	debs := make(map[string]*dagger.File, len(packages))

	for i, v := range args.PackageInputOpts.Packages {
		var (
			opts    = TarOptsFromFileName(v)
			name    = filepath.Base(strings.TrimPrefix(strings.ReplaceAll(v, ".tar.gz", ".deb"), "file://"))
			fpmArgs = []string{
				"fpm",
				"--input-type=dir",
				"--chdir=/pkg",
				"--output-type=deb",
				"--vendor=\"Grafana Labs\"",
				"--name=grafana",
				"--description=Grafana",
				"--url=https://grafana.com",
				"--maintainer=contact@grafana.com",
				"--config-files=/etc/default/grafana-server",
				"--config-files=/usr/lib/systemd/system/grafana-server.service",
				"--config-files=/etc/init.d/grafana-server",
				"--after-install=/src/packaging/deb/control/postinst",
				fmt.Sprintf("--version=%s", opts.Version),
				fmt.Sprintf("--package=%s", "/src/"+name),
				"--depends=adduser",
				"--depends=libfontconfig1",
				"--deb-no-default-config-files", // Help text: Do not add all files in /etc as configuration files by default for Debian packages. (default: false)
			}
		)

		if arch := executil.DebianArch(opts.Distro); arch != "" {
			fpmArgs = append(fpmArgs, fmt.Sprintf("--architecture=%s", arch))
		}
		// TODO: The prerm script was added in v9.5.0; we need to add this flag on versions later than 9.5.0
		// "--before-remove=/src/packaging/deb/control/prerm",

		if !opts.IsEnterprise {
			fpmArgs = append(fpmArgs, "--license=agpl3")
		}

		// The last fpm arg which is required to say, "use the PWD to build the package".
		fpmArgs = append(fpmArgs, ".")

		var (
			// fpm is going to create us a deb package that is going to essentially rsync the folders from the package into the filesystem.
			// These paths are the paths where grafana package contents will be placed.
			packagePaths = []string{
				"/pkg/usr/sbin",
				"/pkg/usr/share",
				// init.d scripts are service management scripts that start/stop/restart/enable the grafana service without systemd.
				// these are likely to be deprecated as systemd is now the default pretty much everywhere.
				"/pkg/etc/init.d",
				// /etc/default holds default environment variables for the grafana-server service
				"/pkg/etc/default",
				// /etc/grafana is empty in the installation, but is set up by the postinstall script and must be created first.
				"/pkg/etc/grafana",
				// these are our systemd unit files that allow systemd to start/stop/restart/enable the grafana service.
				"/pkg/usr/lib/systemd/system",
			}
		)

		container := containers.FPMContainer(d).
			WithEnvVariable("CACHE", "1").
			WithFile("/src/grafana.tar.gz", packages[i]).
			WithExec([]string{"tar", "-xvf", "/src/grafana.tar.gz", "-C", "/src"}).
			WithExec([]string{"ls", "-al", "/src"}).
			WithExec(append([]string{"mkdir", "-p"}, packagePaths...)).
			// the "wrappers" scripts are the same as grafana-cli/grafana-server but with some extra shell commands before/after execution.
			WithExec([]string{"cp", "/src/packaging/wrappers/grafana-server", "/src/packaging/wrappers/grafana-cli", "/pkg/usr/sbin"}).
			WithExec([]string{"cp", "/src/packaging/deb/default/grafana-server", "/pkg/etc/default/grafana-server"}).
			WithExec([]string{"cp", "/src/packaging/deb/init.d/grafana-server", "/pkg/etc/init.d/grafana-server"}).
			WithExec([]string{"cp", "-r", "/src/packaging/deb/systemd/grafana-server.service", "/pkg/usr/lib/systemd/system/grafana-server.service"}).
			WithExec([]string{"cp", "-r", "/src", "/pkg/usr/share/grafana"}).
			WithExec(fpmArgs)

		debs[name] = container.File("/src/" + name)
	}

	for k, v := range debs {
		dst := strings.Join([]string{args.PublishOpts.Destination, k}, "/")
		if err := containers.PublishFile(ctx, d, v, args.PublishOpts, dst); err != nil {
			return err
		}
	}
	return nil
}
