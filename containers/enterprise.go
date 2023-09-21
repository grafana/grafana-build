package containers

import (
	"dagger.io/dagger"
)

const BusyboxImage = "busybox:1.36"

func InitializeEnterprise(d *dagger.Client, grafana *dagger.Directory, enterprise *dagger.Directory) *dagger.Directory {
	hash := d.Container().From(GitImage).
		WithDirectory("/src/grafana-enterprise", enterprise).
		WithWorkdir("/src/grafana-enterprise").
		WithEntrypoint([]string{}).
		WithExec([]string{"/bin/sh", "-c", "git rev-parse --short HEAD > .enterprise-commit"}).
		File("/src/grafana-enterprise/.enterprise-commit")

	// Initializes Grafana Enterprise in the Grafana directory
	g := d.Container().From(BusyboxImage).
		WithDirectory("/src/grafana", grafana).
		WithDirectory("/src/grafana-enterprise", enterprise).
		WithWorkdir("/src/grafana-enterprise").
		WithFile("/src/grafana/.enterprise-commit", hash).
		WithExec([]string{"/bin/sh", "build.sh"}).
		WithExec([]string{"cp", "LICENSE", "../grafana"}).
		WithExec([]string{"cat", "../grafana/.enterprise-commit"}).
		Directory("/src/grafana")

	return g
}
