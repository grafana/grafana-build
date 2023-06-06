package containers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/cliutil"
)

// PublishOpts fields are selectively used based on the protocol field of the destination.
// Be sure to fill out the applicable fields (or all of them) when calling a 'Publish' func.
type PublishOpts struct {
	// Destination is any URL to publish an artifact(s) to.
	// Examples:
	// * '/tmp/package.tar.gz'
	// * 'file:///tmp/package.tar.gz'
	// * 'gcs://bucket/package.tar.gz'
	Destination string

	// Checksum defines if the PublishFile function should also produce / publish a checksum of the given `*dagger.File'
	Checksum bool
}

func PublishOptsFromFlags(c cliutil.CLIContext) *PublishOpts {
	return &PublishOpts{
		Destination: c.String("destination"),
		Checksum:    c.Bool("checksum"),
	}
}

var ErrorUnrecognizedScheme = errors.New("unrecognized scheme")

func publishLocalFile(ctx context.Context, file *dagger.File, dst string) error {
	if _, err := file.Export(ctx, dst); err != nil {
		return err
	}

	return nil
}

func publishGCSFile(ctx context.Context, d *dagger.Client, file *dagger.File, opts *GCPOpts, destination string) error {
	auth := GCSAuth(d, opts)
	uploader, err := GCSUploadFile(d, GoogleCloudImage, auth, file, destination)
	if err != nil {
		return err
	}

	if err := ExitError(ctx, uploader); err != nil {
		return err
	}

	return nil
}

type PublishFileOpts struct {
	File        *dagger.File
	PublishOpts *PublishOpts
	GCPOpts     *GCPOpts
	Destination string
}

// PublishFile publishes the *dagger.File to the specified location. If the destination involves a remote URL or authentication in some way, that information should be populated in the
// `opts *PublishOpts` argument.
func PublishFile(ctx context.Context, d *dagger.Client, opts *PublishFileOpts) ([]string, error) {
	var (
		destination = opts.Destination
		file        = opts.File
		publishOpts = opts.PublishOpts
		gcpOpts     = opts.GCPOpts
	)
	// a map of 'destination' to 'file'
	files := map[string]*dagger.File{
		destination: file,
	}
	if publishOpts.Checksum {
		name := destination + ".sha256"
		log.Println("Checksum is enabled, creating checksum", name)
		files[name] = d.Container().
			From("busybox").
			WithFile("/src/file", file).
			WithExec([]string{"/bin/sh", "-c", "sha256sum /src/file | awk '{print $1}' > /src/file.sha256"}).
			File("/src/file.sha256")
	}

	for dst, f := range files {
		log.Println("Publishing", dst)
		u, err := url.Parse(dst)
		if err != nil {
			// If the destination URL is not a URL then we can assume that it's just a filepath.
			if err := publishLocalFile(ctx, f, dst); err != nil {
				return nil, err
			}
			continue
		}

		switch u.Scheme {
		case "file", "fs":
			dst := strings.TrimPrefix(u.String(), u.Scheme+"://")
			if err := publishLocalFile(ctx, f, dst); err != nil {
				return nil, err
			}
		case "gs":
			if err := publishGCSFile(ctx, d, f, gcpOpts, dst); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("%w: '%s'", ErrorUnrecognizedScheme, u.Scheme)
		}
	}

	out := make([]string, len(files))
	i := 0
	for k := range files {
		out[i] = k
		i++
	}

	return out, nil
}
