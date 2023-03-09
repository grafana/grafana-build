package containers

import "dagger.io/dagger"

const BusyboxImage = "busybox:1.36"

func InitializeEnterprise(d *dagger.Client, grafana *dagger.Directory, enterprise *dagger.Directory) *dagger.Directory {
	return d.Container().From(BusyboxImage).
		WithDirectory("/src/grafana", grafana).
		WithDirectory("/src/grafana-enterprise", enterprise).
		WithWorkdir("/src/grafana-enterprise").
		WithExec([]string{"/bin/sh", "build.sh"}).
		Directory("/src/grafana")
}
