package containers

import (
	"encoding/base64"
	"log"

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

func RPMContainer(d *dagger.Client, opts *GPGOpts) *dagger.Container {
	container := FPMContainer(d).
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-yq", "rpm"})
	if !opts.Sign {
		return container
	}
	var gpgPublicKeyBase64Secret, gpgPrivateKeyBase64Secret *dagger.Secret
	if decodedGPGPublicKeyBase64Secret, err := base64.StdEncoding.DecodeString(opts.GPGPublicKeyBase64); err == nil {
		gpgPublicKeyBase64Secret = d.SetSecret("gpg-public-key", string(decodedGPGPublicKeyBase64Secret))
	} else {
		log.Printf("gpg-public-key-base64 cannot be decoded %s", err.Error())
	}
	if decodedGPGPrivateKeyBase64Secret, err := base64.StdEncoding.DecodeString(opts.GPGPrivateKeyBase64); err == nil {
		gpgPrivateKeyBase64Secret = d.SetSecret("gpg-private-key", string(decodedGPGPrivateKeyBase64Secret))
	} else {
		log.Printf("gpg-private-key-base64 cannot be decoded %s", err.Error())
	}
	gpgPassphraseBase64Secret := d.SetSecret("gpg-passphrase-base64", opts.GPGPassphraseBase64)
	return container.
		WithExec([]string{"apt-get", "install", "-yq", "gnupg2"}).
		WithExec([]string{"mkdir", "-p", "/root/.rpmdb/privkeys"}).
		WithExec([]string{"mkdir", "-p", "/root/.rpmdb/passkeys"}).
		WithExec([]string{"mkdir", "-p", "/root/.rpmdb/pubkeys"}).
		WithSecretVariable("GPG_PUBLIC_KEY", gpgPublicKeyBase64Secret).
		WithSecretVariable("GPG_PRIVATE_KEY", gpgPrivateKeyBase64Secret).
		WithSecretVariable("GPG_PASSPHRASE", gpgPassphraseBase64Secret).
		WithExec([]string{"/bin/sh", "-c", "echo \"$GPG_PRIVATE_KEY\" > /root/.rpmdb/privkeys/grafana.key"}).
		WithExec([]string{"/bin/sh", "-c", "echo \"$GPG_PASSPHRASE\" > /root/.rpmdb/passkeys/grafana.key"}).
		WithExec([]string{"/bin/sh", "-c", "echo \"$GPG_PUBLIC_KEY\" > /root/.rpmdb/pubkeys/grafana.key"}).
		WithNewFile("/root/.rpmmacros", dagger.ContainerWithNewFileOpts{
			Permissions: 0400,
			Contents:    RPMMacros,
		}).
		WithExec([]string{"gpg", "--batch", "--yes", "--no-tty", "--allow-secret-key-import", "--import", "/root/.rpmdb/privkeys/grafana.key"})
}
