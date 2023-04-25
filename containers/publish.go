package containers

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	// * 'file:///tmp/package.tar.gz'
	// * 'gcs://bucket/package.tar.gz'
	Destination string

	GCSOpts *GCSOpts
}

var ErrorUnrecognizedScheme = errors.New("unrecognized scheme")

func publishLocalFile(ctx context.Context, file *dagger.File, dst string) error {
	if _, err := file.Export(ctx, strings.TrimPrefix(dst, "file://")); err != nil {
		return err
	}

	return nil
}

func publishGCSFile(ctx context.Context, d *dagger.Client, file *dagger.File, opts *PublishOpts, destination string) error {
	var auth GCPAuthenticator = &GCPInheritedAuth{}
	// The order of operations:
	// 1. Try to use base64 key.
	// 2. Try to use gcp-service-account-key (path to a file).
	// 3. Try mounting $XDG_CONFIG_HOME/gcloud
	if key := opts.GCSOpts.ServiceAccountKeyBase64; key != "" {
		log.Println("Handling GCP authentication using Service account key from base64...")
		secret := d.SetSecret("gcp-sa-key-base64", key)
		// Write key to a file in an alpine container...
		file := d.Container().From("alpine").
			WithSecretVariable("GCP_SERVICE_ACCOUNT_KEY_BASE64", secret).
			WithExec([]string{"/bin/sh", "-c", "echo $GCP_SERVICE_ACCOUNT_KEY_BASE64 | base64 -d > /key.json"}).
			File("/key.json")

		auth = NewGCPServiceAccountWithFile(file)
	} else if key := opts.GCSOpts.ServiceAccountKey; key != "" {
		log.Println("Handling GCP authentication using Service account key from file...")
		auth = NewGCPServiceAccount(key)
	}

	uploader, err := GCSUploadFile(d, GoogleCloudImage, auth, file, destination)
	if err != nil {
		return err
	}

	if err := ExitError(ctx, uploader); err != nil {
		return err
	}

	return nil
}

func PublishFile(ctx context.Context, d *dagger.Client, file *dagger.File, opts *PublishOpts, destination string) error {
	u, err := url.Parse(opts.Destination)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "file", "fs":
		return publishLocalFile(ctx, file, destination)
	case "gs":
		return publishGCSFile(ctx, d, file, opts, destination)
	}

	return fmt.Errorf("%w: '%s'", ErrorUnrecognizedScheme, u.Scheme)
}

func PublishFiles(ctx context.Context, files []*dagger.File, opts *PublishOpts) error {
	return nil
}

func PublishDirectory(ctx context.Context, dir *dagger.Directory, opts *PublishOpts) error {
	return nil
}
