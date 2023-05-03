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

type InstallerOpts struct {
	PackageType     string
	ConfigFiles     [][]string
	AfterInstall    string
	BeforeRemove    string
	Depends         []string
	EnvFolder       string
	ExtraArgs       []string
	AptDependencies []string
}

// Uses the grafana package given by the '--package' argument and creates a installer.
// It accepts publish args, so you can place the file in a local or remote destination.
func PackageInstaller(ctx context.Context, d *dagger.Client, args PipelineArgs, opts InstallerOpts) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts)
	if err != nil {
		return err
	}

	debs := make(map[string]*dagger.File, len(packages))

	for i, v := range args.PackageInputOpts.Packages {
		var (
			tarOpts = TarOptsFromFileName(v)
			name    = filepath.Base(strings.TrimPrefix(strings.ReplaceAll(v, ".tar.gz", fmt.Sprintf(".%s", opts.PackageType)), "file://"))
			fpmArgs = []string{
				"fpm",
				"--input-type=dir",
				"--chdir=/pkg",
				fmt.Sprintf("--output-type=%s", opts.PackageType),
				"--vendor=\"Grafana Labs\"",
				"--name=grafana",
				"--description=Grafana",
				"--url=https://grafana.com",
				"--maintainer=contact@grafana.com",
				fmt.Sprintf("--version=%s", tarOpts.Version),
				fmt.Sprintf("--package=%s", "/src/"+name),
			}
		)

		for _, c := range opts.ConfigFiles {
			fpmArgs = append(fpmArgs, fmt.Sprintf("--config-files=%s", c[1]))
		}

		if opts.AfterInstall != "" {
			fpmArgs = append(fpmArgs, fmt.Sprintf("--after-install=%s", opts.AfterInstall))
		}

		if opts.BeforeRemove != "" {
			fpmArgs = append(fpmArgs, fmt.Sprintf("--before-remove=%s", opts.BeforeRemove))
		}

		for _, d := range opts.Depends {
			fpmArgs = append(fpmArgs, fmt.Sprintf("--depends=%s", d))
		}

		for _, a := range opts.ExtraArgs {
			fpmArgs = append(fpmArgs, a)
		}

		if arch := executil.PackageArch(tarOpts.Distro); arch != "" {
			fpmArgs = append(fpmArgs, fmt.Sprintf("--architecture=%s", arch))
		}

		if !tarOpts.IsEnterprise {
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
				// holds default environment variables for the grafana-server service
				opts.EnvFolder,
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
			WithExec([]string{"cp", "-r", "/src", "/pkg/usr/share/grafana"})

		for _, conf := range opts.ConfigFiles {
			container = container.WithExec(append([]string{"cp", "-r"}, conf...))
		}

		if len(opts.AptDependencies) > 0 {
			container = container.WithExec([]string{"apt-get", "update"})
			for _, dep := range opts.AptDependencies {
				container = container.WithExec([]string{"apt-get", "install", "-yq", dep})
			}
		}

		container = container.WithExec(fpmArgs)

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
