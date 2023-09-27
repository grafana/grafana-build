package pipelines

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

func VersionPayloadFromFileName(name string) *containers.GCOMVersionPayload {
	var (
		opts         = TarOptsFromFileName(name)
		splitVersion = strings.Split(opts.Version, ".")
		stable       = true
		nightly      = false
		beta         = false
	)

	if strings.Contains(opts.Version, "-") {
		stable = false
		beta = true
	}
	if strings.Contains(opts.Version, "nightly") {
		beta = false
		nightly = true
	}

	return &containers.GCOMVersionPayload{
		Version:         opts.Version,
		ReleaseDate:     time.Now().Format(time.RFC3339Nano),
		Stable:          stable,
		Beta:            beta,
		Nightly:         nightly,
		WhatsNewURL:     fmt.Sprintf("https://grafana.com/docs/grafana/next/whatsnew/whats-new-in-v%s-%s/", splitVersion[0], splitVersion[1]),
		ReleaseNotesURL: "https://grafana.com/docs/grafana/next/release-notes/",
	}
}

func PackagePayloadFromFile(ctx context.Context, d *dagger.Client, name string, file *dagger.File, destination string) (*containers.GCOMPackagePayload, error) {
	opts := TarOptsFromFileName(name)
	ext := filepath.Ext(name)
	os, _ := executil.OSAndArch(opts.Distro)
	arch := strings.ReplaceAll(executil.FullArch(opts.Distro), "/", "")

	if ext == "deb" {
		os = "deb"
	}
	if ext == "rpm" {
		os = "rhel"
	}
	if ext == "exe" {
		os = "win-installer"
	}
	if os == "windows" {
		os = "win"
	}

	u, err := url.Parse(destination)
	if err != nil {
		return nil, err
	}

	sha256, err := containers.Sha256(d, file).Contents(ctx)
	if err != nil {
		return nil, err
	}

	return &containers.GCOMPackagePayload{
		OS:     os,
		URL:    fmt.Sprintf("https://dl.grafana.com%s/%s", strings.TrimRight(u.Path, "/"), name),
		Sha256: sha256,
		Arch:   arch,
	}, nil
}

func PublishGCOM(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	var (
		opts        = args.GCOMOpts
		publishOpts = args.PublishOpts
		wg          = &errgroup.Group{}
		sm          = semaphore.NewWeighted(args.ConcurrencyOpts.Parallel)
	)

	packages, err := containers.GetPackages(ctx, d, args.PackageInputOpts, args.GCPOpts)
	if err != nil {
		return err
	}

	// Extract the package(s)
	for i, name := range args.PackageInputOpts.Packages {
		wg.Go(PublishGCOMFunc(ctx, sm, d, opts, name, packages[i], publishOpts.Destination))
	}
	return wg.Wait()
}

func PublishGCOMFunc(ctx context.Context, sm *semaphore.Weighted, d *dagger.Client, opts *containers.GCOMOpts, path string, file *dagger.File, destination string) func() error {
	return func() error {
		name := filepath.Base(path)
		log.Printf("[%s] Attempting to publish package", name)
		log.Printf("[%s] Acquiring semaphore", name)
		if err := sm.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("failed to acquire semaphore: %w", err)
		}
		defer sm.Release(1)
		log.Printf("[%s] Acquired semaphore", name)

		log.Printf("[%s] Building version payload", name)
		versionPayload := VersionPayloadFromFileName(name)

		log.Printf("[%s] Building package payload", name)
		packagePayload, err := PackagePayloadFromFile(ctx, d, name, file, destination)
		if err != nil {
			return fmt.Errorf("[%s] error: %w", name, err)
		}

		log.Printf("[%s] Publishing package", name)
		err = containers.PublishGCOM(ctx, d, versionPayload, packagePayload, opts)
		if err != nil {
			return fmt.Errorf("[%s] error: %w", name, err)
		}

		log.Printf("[%s] Done publishing package", name)
		return nil
	}
}
