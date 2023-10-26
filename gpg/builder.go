package gpg

import (
	"encoding/base64"
	"fmt"
	"time"

	"dagger.io/dagger"
)

const RPMMacros = `
%_signature gpg
%_gpg_path /root/.gnupg
%_gpg_name Grafana
%_gpgbin /usr/bin/gpg2
%__gpg_sign_cmd %{__gpg} gpg \
	--batch --yes --no-armor --pinentry-mode loopback \
	--passphrase-file /root/.rpmdb/passkeys/grafana.key \
	--no-secmem-warning -u "%{_gpg_name}" -sbo %{__signature_filename} \
	%{?_gpg_digest_algo:--digest-algo %{_gpg_digest_algo}} %{__plaintext_filename}
`

type GPGOpts struct {
	Sign                bool
	GPGPrivateKeyBase64 string
	GPGPublicKeyBase64  string
	GPGPassphrase       string
}

func WithGPGOpts(d *dagger.Client, container *dagger.Container, opts *GPGOpts) (*dagger.Container, error) {
	pubKey, err := base64.StdEncoding.DecodeString(opts.GPGPublicKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("gpg-private-key-base64 cannot be decoded %w", err)
	}

	privKey, err := base64.StdEncoding.DecodeString(opts.GPGPrivateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("gpg-private-key-base64 cannot be decoded %w", err)
	}

	var (
		gpgPassphraseSecret = d.SetSecret("gpg-passphrase", opts.GPGPassphrase)
		gpgPrivateKeySecret = d.SetSecret("gpg-private-key", string(privKey))
		gpgPublicKeySecret  = d.SetSecret("gpg-public-key", string(pubKey))
	)

	return container.
		WithEnvVariable("CACHE_DISABLE", time.Now().String()).
		WithMountedSecret("/root/.rpmdb/privkeys/grafana.key", gpgPrivateKeySecret).
		WithMountedSecret("/root/.rpmdb/pubkeys/grafana.key", gpgPublicKeySecret).
		WithMountedSecret("/root/.rpmdb/passkeys/grafana.key", gpgPassphraseSecret).
		WithNewFile("/root/.rpmmacros", dagger.ContainerWithNewFileOpts{
			Permissions: 0400,
			Contents:    RPMMacros,
		}).
		WithExec([]string{"gpg", "--batch", "--yes", "--no-tty", "--allow-secret-key-import", "--import", "/root/.rpmdb/privkeys/grafana.key"}), nil
}
