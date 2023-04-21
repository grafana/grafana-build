package pipelines

import (
	"context"
	"log"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

// PublishPackage creates a package and publishes it to a Google Cloud Storage bucket.
func PublishPackage(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) error {
	packages, err := PackageFiles(ctx, d, src, args)
	if err != nil {
		return err
	}

	var auth containers.GCPAuthenticator = &containers.GCPInheritedAuth{}
	// The order of operations:
	// 1. Try to use base64 key.
	// 2. Try to use gcp-service-account-key (path to a file).
	// 3. Try mounting $XDG_CONFIG_HOME/gcloud
	if key := args.Context.String("gcp-service-account-key-base64"); key != "" {
		log.Println("Handling GCP authentication using Service account key from base64...")
		secret := d.SetSecret("gcp-sa-key-base64", key)
		// Write key to a file in an alpine container...
		file := d.Container().From("alpine").
			WithSecretVariable("GCP_SERVICE_ACCOUNT_KEY_BASE64", secret).
			WithExec([]string{"/bin/sh", "-c", "echo $GCP_SERVICE_ACCOUNT_KEY_BASE64 | base64 -d > /key.json"}).
			File("/key.json")

		auth = containers.NewGCPServiceAccountWithFile(file)
	} else if key := args.Context.String("gcp-service-account-key"); key != "" {
		log.Println("Handling GCP authentication using Service account key from file...")
		auth = containers.NewGCPServiceAccount(key)
	}

	for distro, targz := range packages {
		fn := TarFilename(args.Version, args.BuildID, args.BuildEnterprise, distro)
		dst := strings.Join([]string{args.Context.Path("destination"), fn}, "/")
		log.Println("Writing package", fn, "to", dst)
		uploader, err := containers.GCSUploadFile(d, containers.GoogleCloudImage, auth, targz, dst)
		if err != nil {
			return err
		}

		if err := containers.ExitError(ctx, uploader); err != nil {
			return err
		}
	}

	return nil
}
