package containers

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"dagger.io/dagger"
)

// GCSOpts are options used when uploading artifacts to Google Cloud Storage.
type GCSOpts struct {
	ServiceAccountKey       string
	ServiceAccountKeyBase64 string
}

// PublishOpts fields are selectively used based on the protocol field of the destination.
// Be sure to fill out the applicable fields (or all of them) when calling a 'Publish' func.
type PublishOpts struct {
	// Destination is any URL to publish an artifact(s) to.
	// Examples:
	// * '/tmp/package.tar.gz'
	// * 'file:///tmp/package.tar.gz'
	// * 'gcs://bucket/package.tar.gz'
	Destination string

	GCSOpts *GCSOpts
}

var ErrorUnrecognizedScheme = errors.New("unrecognized scheme")

func publishLocalFile(ctx context.Context, file *dagger.File, dst string) error {
	if _, err := file.Export(ctx, dst); err != nil {
		return err
	}

	return nil
}

func publishGCSFile(ctx context.Context, d *dagger.Client, file *dagger.File, opts *PublishOpts, destination string) error {
	auth := GCSAuth(d, opts.GCSOpts)
	uploader, err := GCSUploadFile(d, GoogleCloudImage, auth, file, destination)
	if err != nil {
		return err
	}

	if err := ExitError(ctx, uploader); err != nil {
		return err
	}

	return nil
}

// PublishFile publishes the *dagger.File to the specified location. If the destination involves a remote URL or authentication in some way, that information should be populated in the
// `opts *PublishOpts` argument.
func PublishFile(ctx context.Context, d *dagger.Client, file *dagger.File, opts *PublishOpts, destination string) error {
	u, err := url.Parse(destination)
	if err != nil {
		// If the destination URL is not a URL then we can assume that it's just a filepath.
		return publishLocalFile(ctx, file, destination)
	}

	switch u.Scheme {
	case "file", "fs":
		dst := strings.TrimPrefix(u.String(), u.Scheme+"://")
		return publishLocalFile(ctx, file, dst)
	case "gs":
		return publishGCSFile(ctx, d, file, opts, destination)
	}

	return fmt.Errorf("%w: '%s'", ErrorUnrecognizedScheme, u.Scheme)
}

func PublishDirectory(ctx context.Context, dir *dagger.Directory, opts *PublishOpts) error {
	return nil
}
