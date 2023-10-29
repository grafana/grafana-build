package pipelines

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/fpm"
	"github.com/grafana/grafana-build/gpg"
)

func WithRPMSignature(ctx context.Context, d *dagger.Client, opts *gpg.GPGOpts, installers map[string]*dagger.File) (map[string]*dagger.File, error) {
	out := make(map[string]*dagger.File, len(installers))

	for dst, file := range installers {
		if filepath.Ext(dst) != ".rpm" {
			log.Println(dst, "is not an rpm, it is", filepath.Ext(dst))
			out[dst] = file
			continue
		}

		container, err := gpg.WithGPGOpts(d, fpm.Builder(d), opts)
		if err != nil {
			return nil, err
		}
		f := container.
			WithEnvVariable("dst", dst).
			WithMountedFile("/src/package.rpm", file).
			WithExec([]string{"rpm", "--addsign", "/src/package.rpm"}).
			WithExec([]string{"/bin/sh", "-c", fmt.Sprintf("rpm --checksig %s | grep -qE 'digests signatures OK|pgp.+OK'", "/src/package.rpm")}).
			File("/src/package.rpm")

		out[dst] = f
	}

	return out, nil
}
