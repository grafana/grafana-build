package pipelines

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	grafanabuild "github.com/grafana/grafana-build"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
	"github.com/grafana/grafana-build/tarfs"
)

const winSWx64URL = "https://github.com/winsw/winsw/releases/download/v2.12.0/WinSW-x64.exe"

// WindowsInstaller uses a debian image with 'nsis' installed to create the Windows installer.
// The WindowsInstaller also uses "winsw", or "Windows Service Wrapper" (https://github.com/winsw/winsw) to handle the status, start, and stop functions of
// the windows service so that Grafana doesn't have to implement it.
func WindowsInstaller(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return err
	}

	// Download the 'winsw' executable
	winsw := d.Container().From("busybox").
		WithExec([]string{"wget", winSWx64URL, "-O", "/grafana-svc.exe"}).
		File("/grafana-svc.exe")

	// Build an MSI installer for each package given
	exes := make(map[string]*dagger.File, len(packages))

	base := d.Container().From("debian:sid").
		WithMountedFile("/src/grafana-svc.exe", winsw).
		WithExec([]string{"apt-get", "update", "-yq"}).
		WithExec([]string{"apt-get", "install", "tar", "nsis"})

	// Write the embedded contents into a gzipped tar archive to mount into the container.
	// I tried mounting each file individually by following the guide here:
	// https://docs.dagger.io/110632/embed-directories/
	// but ran into graphql syntax errors on all images and the rtf file, so this is a workaround.

	// Write the buffer to a tempdir to be mounted
	f, err := os.Create("windows-packaging.tar.gz")
	if err != nil {
		return err
	}

	// Close and remove the file whenever this function is done.
	defer func() {
		n := f.Name()
		if err := f.Close(); err != nil {
			log.Println("Error closing", n, "error:", err)
			return
		}
		if err := os.Remove(n); err != nil {
			log.Println("Error removing file", n, "error:", err)
		}
	}()

	if err := tarfs.Write(f, grafanabuild.WindowsPackaging); err != nil {
		return err
	}

	packaging := d.Host().Directory(filepath.Dir(f.Name())).File(filepath.Base(f.Name()))

	base = base.WithMountedFile("/src/src.tar.gz", packaging).
		WithExec([]string{"tar", "--strip-components=1", "-xzf", "/src/src.tar.gz", "--strip-components=3", "-C", "/src"})

	for i, v := range args.PackageInputOpts.Packages {
		var (
			taropts = TarOptsFromFileName(v)
			name    = DestinationName(v, "exe")
			targz   = packages[i]
		)
		log.Println("Taropts from file name", v, taropts)

		if os, _ := executil.OSAndArch(taropts.Distro); os != "windows" {
			return fmt.Errorf("package '%s' is not a windows package", v)
		}

		var (
			app     = "Grafana"
			slug    = "grafana"
			license = "AGPLv3.rtf" // TODO: we should probably prefer that the installer shows the license.txt in the repo instead.
		)

		if taropts.Edition == "enterprise" {
			app = "Grafana Enterprise"
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

		log.Println("nsis args", nsisArgs)

		exe := base.
			WithFile("/src/grafana.tar.gz", targz).
			WithExec([]string{"mkdir", "-p", "/src/grafana"}).
			WithExec([]string{"tar", "--strip-components=1", "-xzf", "/src/grafana.tar.gz", "-C", "/src/grafana"}).
			WithWorkdir("/src").
			WithExec(nsisArgs).
			File("/src/grafana-setup.exe")

		exes[name] = exe
	}

	for k, v := range exes {
		dst := strings.Join([]string{args.PublishOpts.Destination, k}, "/")
		log.Println(k, v, dst)
		out, err := containers.PublishFile(ctx, d, &containers.PublishFileOpts{
			File:        v,
			Destination: dst,
			PublishOpts: args.PublishOpts,
			GCPOpts:     args.GCPOpts,
		})
		if err != nil {
			return err
		}

		WriteToStdout(out)
	}

	return nil
}
