package exe

import (
	"dagger.io/dagger"
	grafanabuild "github.com/grafana/grafana-build"
	"github.com/grafana/grafana-build/containers"
)

const winSWx64URL = "https://github.com/winsw/winsw/releases/download/v2.12.0/WinSW-x64.exe"

func Builder(d *dagger.Client) (*dagger.Container, error) {
	winsw := d.Container().From("busybox").
		WithExec([]string{"wget", winSWx64URL, "-O", "/grafana-svc.exe"}).
		File("/grafana-svc.exe")

	debian := d.Container().From("debian:sid").
		WithExec([]string{"apt-get", "update", "-yq"}).
		WithExec([]string{"apt-get", "install", "tar", "nsis"})

	builder, err := containers.WithEmbeddedFS(d, debian, "/src", grafanabuild.WindowsPackaging)
	if err != nil {
		return nil, err
	}

	return builder.WithFile("/src/grafana-svc.exe", winsw), nil
}
