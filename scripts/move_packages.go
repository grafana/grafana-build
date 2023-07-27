package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
	"github.com/grafana/grafana-build/pipelines"
)

const (
	// 1: The version (with a v prefix)
	// 2: The "edition". Options: 'oss', 'pro', 'enterprise'.
	// 3: The full name. 'grafana', 'grafana-enterprise', 'grafana-pro
	// 4: The 'ersion', or 'version' without the 'v'.
	// 5: The OS: 'windows', 'linux', 'darwin'
	// 6: The architecture: 'amd64', 'armv6', 'armv7', 'arm64'.
	// 7: -musl, sometimes.
	// 8: '.sha256', sometimes.
	tarGzFormat = "artifacts/downloads%[9]s/%[1]s/%[2]s/release/%[3]s-%[4]s.%[5]s-%[6]s%[7]s.tar.gz%[8]s"
	debFormat   = "artifacts/downloads%[9]s/%[1]s/%[2]s/release/%[3]s_%[4]s_%[6]s.deb%[8]s"
	rpmFormat   = "artifacts/downloads%[9]s/%[1]s/%[2]s/release/%[3]s-%[4]s-1.%[6]s.rpm%[8]s"
	exeFormat   = "artifacts/downloads%[9]s/%[1]s/%[2]s/release/%[3]s_%[4]s_%[6]s.exe%[8]s"

	// 1: ersion
	// 2. name (grafana-oss | grafana-enterprise)
	// 3: '-ubuntu', if set
	// 4: arch
	// 5: '.sha256', if set
	dockerFormat = "artifacts/docker/%[1]s/%[2]s-%[1]s%[3]s-%[4]s.img%[5]s"

	// 1: ersion
	// 2. name (grafana-oss | grafana-enterprise)
	cdnFormat = "artifacts/static-assets/%[2]s/%[1]s/public"

	// 1: ersion
	storybookFormat = "artifacts/storybook/%[1]s"

	// 1: version
	// 2: package name (@grafana-ui-10.0.0.tgz)
	npmFormat = "artifacts/npm/%[1]s/npm-artifacts/%[2]s"

	sha256Ext = ".sha256"
	grafana   = "grafana"
)

// One artifact and be copied to multiple different locations (like armv7 tar.gz packages should be copied to tar.gz and -musl.tar.gz)
type HandlerFunc func(name string) []string

var Handlers = map[string]HandlerFunc{
	".tar.gz":        TarGZHandler,
	".tgz":           NPMHandler,
	".deb":           DebHandler,
	".rpm":           RPMHandler,
	".docker.tar.gz": DockerHandler,
	".exe":           EXEHandler,
	".zip":           ZipHandler,
}

func NPMHandler(name string) []string {
	var (
		version = os.Getenv("DRONE_TAG")
		file    = filepath.Base(name)
	)

	return []string{fmt.Sprintf(npmFormat, version, file)}
}

func ZipHandler(name string) []string {
	files := EXEHandler(strings.ReplaceAll(name, "zip", "exe"))

	for i, v := range files {
		files[i] = strings.ReplaceAll(v, "exe", "zip")
	}

	return files
}

func RPMHandler(name string) []string {
	ext := filepath.Ext(name)

	// If we're copying a sha256 file and not a tar.gz then we want to add .sha256 to the template
	// or just give it emptystring if it's not the sha256 file
	sha256 := ""
	if ext == sha256Ext {
		sha256 = sha256Ext
	}

	n := filepath.Base(name) // Surprisingly still works even with 'gs://' urls
	opts := pipelines.TarOptsFromFileName(strings.ReplaceAll(strings.ReplaceAll(n, sha256Ext, ""), "rpm", "tar.gz"))

	// In grafana-build we just use "" to refer to "oss"
	edition := "oss"
	fullName := grafana
	if opts.Edition != "" {
		edition = opts.Edition
		fullName += "-" + opts.Edition
	}

	goos, arch := executil.OSAndArch(opts.Distro)
	arm := executil.ArchVersion(opts.Distro)
	if arch == "arm" {
		if arm == "7" {
			arch = "armhfp"
		}
	}

	if arch == "arm64" {
		arch = "aarch64"
	}

	if arch == "amd64" {
		arch = "x86_64"
	}

	enterprise2 := ""
	version := opts.Version
	ersion := strings.Replace(strings.TrimPrefix(version, "v"), "-", "~", 1)

	if edition == "pro" {
		// "pro" in this case is called "enterprise2"
		fullName = "grafana-enterprise2"
		edition = "enterprise2"
		// and is in the 'downloads-enterprise2' folder instead of 'downloads'
		enterprise2 = "-enterprise2"
		// and has an period separator {version}.{arch} instead of {version}_{arch}
	}
	dst := fmt.Sprintf(rpmFormat, version, edition, fullName, ersion, goos, arch, edition, sha256, enterprise2)

	return []string{
		dst,
	}
}

func EXEHandler(name string) []string {
	packages := DebHandler(strings.ReplaceAll(name, "exe", "deb"))
	for i, v := range packages {
		v = strings.ReplaceAll(v, "deb", "exe")
		v = strings.ReplaceAll(v, "amd64", "windows-amd64")
		v = strings.ReplaceAll(v, "_", "-")
		v = strings.ReplaceAll(v, "~", "-")
		v = strings.ReplaceAll(v, "-windows", ".windows")
		packages[i] = v
	}

	return packages
}

func DebHandler(name string) []string {
	ext := filepath.Ext(name)
	format := debFormat

	// If we're copying a sha256 file and not a tar.gz then we want to add .sha256 to the template
	// or just give it emptystring if it's not the sha256 file
	sha256 := ""
	if ext == sha256Ext {
		sha256 = sha256Ext
	}

	n := filepath.Base(name) // Surprisingly still works even with 'gs://' urls
	opts := pipelines.TarOptsFromFileName(strings.ReplaceAll(strings.ReplaceAll(n, sha256Ext, ""), "deb", "tar.gz"))

	// In grafana-build we just use "" to refer to "oss"
	edition := "oss"
	fullName := grafana
	version := opts.Version
	ersion := strings.TrimPrefix(version, "v")
	ersion = strings.Replace(ersion, "-", "~", 1)
	enterprise2 := ""
	if opts.Edition != "" {
		edition = opts.Edition
		fullName += "-" + opts.Edition
		if edition == "pro" {
			// "pro" in this case is called "enterprise2"
			fullName = "grafana-enterprise2"
			edition = "enterprise2"
			// and is in the 'downloads-enterprise2' folder instead of 'downloads'
			enterprise2 = "-enterprise2"
		}
	}

	names := []string{fullName}
	goos, arch := executil.OSAndArch(opts.Distro)
	arm := executil.ArchVersion(opts.Distro)
	if arch == "arm" {
		if arm == "7" {
			arch = "armhf"
		}
		// If we're building for arm then we also copy the same thing, but with the name '-rpi'. for osme reason?
		names = []string{fullName, fullName + "-rpi"}
	}

	dst := []string{}
	for _, n := range names {
		dst = append(dst, fmt.Sprintf(format, opts.Version, edition, n, ersion, goos, arch, edition, sha256, enterprise2))
	}

	return dst
}

func TarGZHandler(name string) []string {
	ext := filepath.Ext(name)

	// If we're copying a sha256 file and not a tar.gz then we want to add .sha256 to the template
	// or just give it emptystring if it's not the sha256 file
	sha256 := ""
	if ext == sha256Ext {
		sha256 = sha256Ext
	}

	n := filepath.Base(name) // Surprisingly still works even with 'gs://' urls
	opts := pipelines.TarOptsFromFileName(strings.ReplaceAll(n, sha256Ext, ""))

	// In grafana-build we just use "" to refer to "oss"
	edition := "oss"
	fullName := grafana
	version := opts.Version
	ersion := strings.TrimPrefix(version, "v")
	enterprise2 := ""
	if opts.Edition != "" {
		edition = opts.Edition
		fullName += "-" + opts.Edition
		if edition == "pro" {
			enterprise2 = "-enterprise2"
			fullName = "grafana-enterprise2"
			edition = "enterprise2"
		}
	}

	libc := []string{""}
	goos, arch := executil.OSAndArch(opts.Distro)

	if arch == "arm64" || arch == "arm" || arch == "amd64" && goos == "linux" {
		libc = []string{"", "-musl"}
	}

	arm := executil.ArchVersion(opts.Distro)
	if arch == "arm" {
		arch += "v" + arm
		// I guess we don't create an arm-6-musl?
		if arm == "6" {
			libc = []string{""}
		}
	}

	dst := []string{}
	for _, m := range libc {
		dst = append(dst, fmt.Sprintf(tarGzFormat, opts.Version, edition, fullName, ersion, goos, arch, m, sha256, enterprise2))
	}

	return dst
}

func DockerHandler(name string) []string {
	ext := filepath.Ext(name)

	// If we're copying a sha256 file and not a tar.gz then we want to add .sha256 to the template
	// or just give it emptystring if it's not the sha256 file
	sha256 := ""
	if ext == sha256Ext {
		sha256 = sha256Ext
	}

	n := filepath.Base(name) // Surprisingly still works even with 'gs://' urls

	// try to get .ubuntu.docker.tar.gz.sha256 / .ubuntu.docker.tar.gz / docker.tar.gz to all just end in 'tar.gz'
	normalized := strings.ReplaceAll(n, sha256Ext, "")
	normalized = strings.ReplaceAll(normalized, ".ubuntu", "")
	normalized = strings.ReplaceAll(normalized, ".docker", "")

	opts := pipelines.TarOptsFromFileName(normalized)

	// In grafana-build we just use "" to refer to "oss"
	edition := "oss"
	fullName := grafana
	if opts.Edition != "" {
		edition = opts.Edition
		if edition == "pro" {
			edition = "enterprise2"
		}
	}

	fullName += "-" + edition
	ubuntu := ""
	if strings.Contains(name, "ubuntu") {
		ubuntu = "-ubuntu"
	}

	_, arch := executil.OSAndArch(opts.Distro)
	if arch == "arm" {
		arch += "v" + executil.ArchVersion(opts.Distro)
	}
	return []string{
		fmt.Sprintf(dockerFormat, strings.TrimPrefix(opts.Version, "v"), fullName, ubuntu, arch, sha256),
	}
}

func CDNHandler(name string) []string {
	n := filepath.Base(strings.ReplaceAll(name, "/public", ".tar.gz")) // Surprisingly still works even with 'gs://' urls

	opts := pipelines.TarOptsFromFileName(n)

	// In grafana-build we just use "" to refer to "oss"
	edition := "oss"
	fullName := grafana
	if opts.Edition != "" {
		edition = opts.Edition
	}

	fullName += "-" + edition

	names := []string{
		fmt.Sprintf(cdnFormat, strings.TrimPrefix(opts.Version, "v"), fullName),
	}

	if edition == "oss" {
		names = append(names, fmt.Sprintf(cdnFormat, strings.TrimPrefix(opts.Version, "v"), grafana))
	}

	return names
}

func StorybookHandler(name string) []string {
	n := filepath.Base(strings.ReplaceAll(name, "/storybook", ".tar.gz")) // Surprisingly still works even with 'gs://' urls
	opts := pipelines.TarOptsFromFileName(n)

	names := []string{
		fmt.Sprintf(storybookFormat, strings.TrimPrefix(opts.Version, "v")),
	}

	return names
}

// A hopefully temporary script that prints the gsutil commands that will move these artifacts into the location where they were expected previously.
// Just pipe this into bash or exec or whatever to do the actual copying.
// Run without redirecting stdout to verify the operations.
func main() {
	prefix := os.Args[1]

	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		panic(err)
	}

	var (
		scanner       = bufio.NewScanner(os.Stdin)
		authenticator = containers.GCSAuth(client, &containers.GCPOpts{
			ServiceAccountKeyBase64: os.Getenv("GCP_KEY_BASE64"),
		})

		container = client.Container().From("google/cloud-sdk:alpine")
	)

	if c, err := authenticator.Authenticate(client, container); err == nil {
		container = c
	} else {
		panic(err)
	}

	for scanner.Scan() {
		var (
			name = scanner.Text()
			ext  = filepath.Ext(name)
		)

		// sha256 extensions should be handled the same way what precedes the extension
		if ext == sha256Ext {
			ext = filepath.Ext(strings.ReplaceAll(name, sha256Ext, ""))
		}

		// tar.gz extensions can also have docker.tar.gz so we need to make sure we don't skip that
		if ext == ".gz" {
			ext = ".tar.gz"
			if filepath.Ext(strings.ReplaceAll(name, ext, "")) == ".docker" {
				ext = ".docker.tar.gz"
			}
		}
		handler := Handlers[ext]
		if ext == "" {
			destinations := make([]string, 0)
			if filepath.Base(name) == "public" {
				destinations = CDNHandler(name)
			}
			if filepath.Base(name) == "storybook" {
				destinations = StorybookHandler(name)
			}
			for _, v := range destinations {
				dir := filepath.Join(prefix, filepath.Dir(v))
				v := filepath.Join(prefix, v)

				log.Println("Creating dir", dir)
				if err := os.MkdirAll(dir, 0700); err != nil {
					panic(err)
				}
				log.Println("Copying", name, "to", v)
				//nolint:gosec
				cmd := exec.Command("cp", "-r", strings.TrimPrefix(name, "file://"), v)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					panic(err)
				}
				//container = container.WithExec([]string{"gcloud", "storage", "cp", "-r", name, v})
			}
			continue
		}

		destinations := handler(name)
		for _, v := range destinations {
			dir := filepath.Join(prefix, filepath.Dir(v))
			v := filepath.Join(prefix, v)
			log.Println("Creating directory", dir)
			if err := os.MkdirAll(dir, 0700); err != nil {
				panic(err)
			}
			log.Println("Copying", name, "to", dir)
			//nolint:gosec
			cmd := exec.Command("cp", strings.TrimPrefix(name, "file://"), v)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				panic(err)
			}
			//container = container.WithExec([]string{"gcloud", "storage", "cp", name, v})
		}
	}

	log.Println("Copying", prefix, "to gcs")
	dst := os.Getenv("DESTINATION")
	container = container.WithMountedDirectory("dist", client.Host().Directory(prefix)).
		WithExec([]string{"gcloud", "storage", "cp", "-r", "/dist/artifacts", dst})

	stdout, err := container.Stdout(ctx)
	if err != nil {
		panic(err)
	}

	stderr, err := container.Stdout(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Fprint(os.Stdout, stdout)
	fmt.Fprint(os.Stderr, stderr)
}
