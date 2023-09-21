package containers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"dagger.io/dagger"
)

// PackagePublish returns a docker container with everything set up that is needed to publish deb and rpm packages.
func PackagePublish(ctx context.Context, d *dagger.Client, opts *PackagePublishOpts, packageType string) (*dagger.Container, error) {
	pubKey, err := base64.StdEncoding.DecodeString(opts.GPGPublicKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("gpg-private-key-base64 cannot be decoded %w", err)
	}

	privKey, err := base64.StdEncoding.DecodeString(opts.GPGPrivateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("gpg-private-key-base64 cannot be decoded %w", err)
	}

	serviceAccountKey, err := base64.StdEncoding.DecodeString(opts.ServiceAccountKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("gcp-service-account-key-base64 cannot be decoded %w", err)
	}

	var (
		targetBucket          = strings.TrimPrefix(opts.Destination, "gs://")
		accessKeyIdSecret     = d.SetSecret("access-key-id", opts.AccessKeyId)
		secretAccessKeySecret = d.SetSecret("secret-access-key", opts.SecretAccessKey)
		gpgPassphraseSecret   = d.SetSecret("gpg-passphrase", opts.GPGPassphrase)
		gpgPrivateKeySecret   = d.SetSecret("gpg-private-key", string(privKey))
		gpgPublicKeySecret    = d.SetSecret("gpg-public-key", string(pubKey))
		serviceAccountSecret  = d.SetSecret("service-account", string(serviceAccountKey))
	)

	return d.Container().From("us.gcr.io/kubernetes-dev/package-publish").
		WithSecretVariable("ACCESS_KEY_ID", accessKeyIdSecret).
		WithSecretVariable("SECRET_ACCESS_KEY", secretAccessKeySecret).
		WithSecretVariable("GPG_PASSPHRASE", gpgPassphraseSecret).
		WithSecretVariable("GPG_PRIVATE_KEY", gpgPrivateKeySecret).
		WithSecretVariable("GPG_PUBLIC_KEY", gpgPublicKeySecret).
		WithSecretVariable("SERVICE_ACCOUNT_JSON", serviceAccountSecret).
		WithEnvVariable("TARGET_BUCKET", targetBucket).
		WithEnvVariable("REPLACE_EXISTING", strconv.FormatBool(opts.ReplaceExisting)).
		WithEnvVariable("DEB_REMOVE_PACKAGES", strconv.FormatBool(opts.RemovePackages)).
		WithEnvVariable("PACKAGE_PATH", strings.Join(opts.Packages, ",")).
		WithEnvVariable("PACKAGE_TYPE", packageType).
		WithExec(nil).
		Sync(ctx)
}
