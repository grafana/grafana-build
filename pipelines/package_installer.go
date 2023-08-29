package pipelines

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
	"github.com/grafana/grafana-build/versions"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type InstallerOpts struct {
	NameOverride string
	PackageType  string
	ConfigFiles  [][]string
	AfterInstall string
	BeforeRemove string
	Depends      []string
	EnvFolder    string
	ExtraArgs    []string
	RPMSign      bool
	Container    *dagger.Container
}

type InstallerFunc func(ctx context.Context, d *dagger.Client, args PipelineArgs, opts InstallerOpts) error

func PublishInstallers(ctx context.Context, d *dagger.Client, args PipelineArgs, packages map[string]*dagger.File) error {
	var (
		wg = &errgroup.Group{}
		sm = semaphore.NewWeighted(args.ConcurrencyOpts.Parallel)
	)

	for dst, file := range packages {
		wg.Go(PublishFileFunc(ctx, sm, d, &containers.PublishFileOpts{
			Destination: dst,
			File:        file,
			GCPOpts:     args.GCPOpts,
			PublishOpts: args.PublishOpts,
		}))
	}

	return wg.Wait()
}

// Uses the grafana package given by the '--package' argument and creates a installer.
// It accepts publish args, so you can place the file in a local or remote destination.
func PackageInstaller(ctx context.Context, d *dagger.Client, args PipelineArgs, opts InstallerOpts) (map[string]*dagger.File, error) {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return nil, err
	}

	installers := make(map[string]*dagger.File, len(packages))

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
				"--url=https://grafana.com",
				"--maintainer=contact@grafana.com",
				fmt.Sprintf("--version=%s", strings.TrimPrefix(tarOpts.Version, "v")),
				fmt.Sprintf("--package=%s", "/src/"+name),
			}

			vopts = versions.OptionsFor(tarOpts.Version)
		)

		// If this is a debian installer and this version had a prerm script (introduced in v9.5)...
		// TODO: this logic means that rpms can't also have a beforeremove. Not important at the moment because it's static (in pipelines/rpm.go) and it doesn't have beforeremove set.
		if vopts.DebPreRM.IsSet && vopts.DebPreRM.Value && opts.PackageType == "deb" {
			if opts.BeforeRemove != "" {
				fpmArgs = append(fpmArgs, fmt.Sprintf("--before-remove=%s", opts.BeforeRemove))
			}
		}

		// These paths need to be absolute when installed on the machine and not the package structure.
		for _, c := range opts.ConfigFiles {
			fpmArgs = append(fpmArgs, fmt.Sprintf("--config-files=%s", strings.TrimPrefix(c[1], "/pkg")))
		}

		if opts.AfterInstall != "" {
			fpmArgs = append(fpmArgs, fmt.Sprintf("--after-install=%s", opts.AfterInstall))
		}

		for _, d := range opts.Depends {
			fpmArgs = append(fpmArgs, fmt.Sprintf("--depends=%s", d))
		}

		fpmArgs = append(fpmArgs, opts.ExtraArgs...)

		if arch := executil.PackageArch(tarOpts.Distro); arch != "" {
			fpmArgs = append(fpmArgs, fmt.Sprintf("--architecture=%s", arch))
		}

		packageName := "grafana"
		// Honestly we don't care about making fpm installers for non-enterprise or non-grafana flavors of grafana
		if tarOpts.Edition == "enterprise" || tarOpts.Edition == "pro" {
			packageName = fmt.Sprintf("grafana-%s", tarOpts.Edition)
			fpmArgs = append(fpmArgs, "--description=\"Grafana Enterprise\"")
			fpmArgs = append(fpmArgs, "--conflicts=grafana")
		} else {
			fpmArgs = append(fpmArgs, "--description=Grafana")
			fpmArgs = append(fpmArgs, "--license=AGPLv3")
		}

		if n := opts.NameOverride; n != "" {
			packageName = n
		}

		fpmArgs = append(fpmArgs, fmt.Sprintf("--name=%s", packageName))

		// The last fpm arg which is required to say, "use the PWD to build the package".
		fpmArgs = append(fpmArgs, ".")

		var (
			// fpm is going to create us a package that is going to essentially rsync the folders from the package into the filesystem.
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

		container := opts.Container.
			WithFile("/src/grafana.tar.gz", packages[i]).
			WithEnvVariable("XZ_DEFAULTS", "-T0").
			WithExec([]string{"tar", "--exclude=storybook", "--strip-components=1", "-xvf", "/src/grafana.tar.gz", "-C", "/src"}).
			WithExec([]string{"rm", "/src/grafana.tar.gz"})

		container = container.
			WithExec(append([]string{"mkdir", "-p"}, packagePaths...)).
			// the "wrappers" scripts are the same as grafana-cli/grafana-server but with some extra shell commands before/after execution.
			WithExec([]string{"cp", "/src/packaging/wrappers/grafana-server", "/src/packaging/wrappers/grafana-cli", "/pkg/usr/sbin"}).
			WithExec([]string{"cp", "-r", "/src", "/pkg/usr/share/grafana"})

		for _, conf := range opts.ConfigFiles {
			container = container.WithExec(append([]string{"cp", "-r"}, conf...))
		}

		container = container.WithExec(fpmArgs)
		dst := strings.Join([]string{args.PublishOpts.Destination, strings.ReplaceAll(name, tarOpts.Name, packageName)}, "/")
		installers[dst] = container.File("/src/" + name)
	}

	return installers, nil
}

type SyncWriter struct {
	Writer io.Writer

	mutex *sync.Mutex
}

func NewSyncWriter(w io.Writer) *SyncWriter {
	return &SyncWriter{
		Writer: w,
		mutex:  &sync.Mutex{},
	}
}

func (w *SyncWriter) Write(b []byte) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.Writer.Write(b)
}

var Stdout = NewSyncWriter(os.Stdout)

func PublishFileFunc(ctx context.Context, sm *semaphore.Weighted, d *dagger.Client, opts *containers.PublishFileOpts) func() error {
	return func() error {
		log.Printf("[%s] Attempting to publish file", opts.Destination)
		log.Printf("[%s] Acquiring semaphore", opts.Destination)
		if err := sm.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("failed to acquire semaphore: %w", err)
		}
		defer sm.Release(1)
		log.Printf("[%s] Acquired semaphore", opts.Destination)

		log.Printf("[%s] Publishing file", opts.Destination)
		out, err := containers.PublishFile(ctx, d, opts)
		if err != nil {
			return fmt.Errorf("[%s] error: %w", opts.Destination, err)
		}
		log.Printf("[%s] Done publishing file", opts.Destination)

		fmt.Fprintln(Stdout, strings.Join(out, "\n"))
		return nil
	}
}

func PublishDirFunc(ctx context.Context, sm *semaphore.Weighted, d *dagger.Client, dir *dagger.Directory, opts *containers.GCPOpts, dst string) func() error {
	return func() error {
		log.Printf("[%s] Attempting to publish file", dst)
		log.Printf("[%s] Acquiring semaphore", dst)
		if err := sm.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("failed to acquire semaphore: %w", err)
		}
		defer sm.Release(1)
		log.Printf("[%s] Acquired semaphore", dst)

		log.Printf("[%s] Publishing file", dst)
		out, err := containers.PublishDirectory(ctx, d, dir, opts, dst)
		if err != nil {
			return fmt.Errorf("[%s] error: %w", dst, err)
		}
		log.Printf("[%s] Done publishing file", dst)

		fmt.Fprintln(Stdout, out)
		return nil
	}
}
