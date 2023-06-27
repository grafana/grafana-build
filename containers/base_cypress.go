package containers

import (
	"fmt"
	"strings"

	"dagger.io/dagger"
)

func CypressImage(version string) string {
	return fmt.Sprintf("cypress/base:%s", strings.TrimPrefix(strings.TrimSpace(version), "v"))
}

// CypressContainer returns a docker container with everything set up that is needed to build or run e2e tests.
func CypressContainer(d *dagger.Client, base string) *dagger.Container {
	container := d.Container(dagger.ContainerOpts{
		Platform: "linux/amd64",
	}).From(base)

	// Install dependencies
	container = container.
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{
			"apt-get", "install", "-y",
			"fonts-liberation",
			"git",
			"libcurl4",
			"libcurl3-gnutls",
			"libcurl3-nss",
			"xdg-utils",
			"wget",
			"curl",
		})

	// Clean-up
	container = container.
		WithExec([]string{"rm", "-rf", "/var/lib/apt/lists/*"}).
		WithExec([]string{"apt-get", "clean"})

	// install libappindicator3-1 - not included with Debian 11
	container = container.
		WithExec([]string{"wget", "--no-verbose", "-O", "libindicator3-7_0.5.0-4_amd64.deb", "http://ftp.us.debian.org/debian/pool/main/libi/libindicator/libindicator3-7_0.5.0-4_amd64.deb"}).
		WithExec([]string{"wget", "--no-verbose", "-O", "libappindicator3-1_0.4.92-7_amd64.deb", "http://ftp.us.debian.org/debian/pool/main/liba/libappindicator/libappindicator3-1_0.4.92-7_amd64.deb"}).
		WithExec([]string{"apt-get", "install", "-f", "-y", "./libindicator3-7_0.5.0-4_amd64.deb"}).
		WithExec([]string{"apt-get", "install", "-f", "-y", "./libappindicator3-1_0.4.92-7_amd64.deb"}).
		WithExec([]string{"rm", "-f", "libindicator3-7_0.5.0-4_amd64.deb"}).
		WithExec([]string{"rm", "-f", "libappindicator3-1_0.4.92-7_amd64.deb"})

	// install Chrome browser
	container = container.
		WithExec([]string{"wget", "--no-verbose", "-O", "google-chrome-stable_current_amd64.deb", "http://dl.google.com/linux/chrome/deb/pool/main/g/google-chrome-stable/google-chrome-stable_107.0.5304.121-1_amd64.deb"}).
		WithExec([]string{"apt-get", "install", "-f", "-y", "./google-chrome-stable_current_amd64.deb"}).
		WithExec([]string{"rm", "-f", "google-chrome-stable_current_amd64.deb"})

	// Set env
	container = container.
		WithEnvVariable("DBUS_SESSION_BUS_ADDRESS", "/dev/null").
		WithEnvVariable("TERM", "xterm").
		WithEnvVariable("npm_config_loglevel", "warn").
		WithEnvVariable("npm_config_unsafe_perm", "true")

	return container
}
