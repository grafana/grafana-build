package containers

import "dagger.io/dagger"

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
	gpgPublicKeySecret := d.SetSecret("gpg-public-key", opts.GPGPublicKey)
	gpgPrivateKeySecret := d.SetSecret("gpg-private-key", opts.GPGPrivateKey)
	gpgPassphraseSecret := d.SetSecret("gpg-passphrase", opts.GPGPassphrase)
	return container.
		WithSecretVariable("GPG_PUBLIC_KEY", gpgPublicKeySecret).
		WithSecretVariable("GPG_PRIVATE_KEY", gpgPrivateKeySecret).
		WithSecretVariable("GPG_PASSPHRASE", gpgPassphraseSecret).
		WithExec([]string{"apt-get", "install", "-yq", "gnupg2"}).
		WithExec([]string{"mkdir", "-p", "/root/.rpmdb/privkeys"}).
		WithExec([]string{"mkdir", "-p", "/root/.rpmdb/passkeys"}).
		WithExec([]string{"mkdir", "-p", "/root/.rpmdb/pubkeys"}).
		WithExec([]string{"/bin/sh", "-c", "echo \"$GPG_PRIVATE_KEY\" | base64 -d > /root/.rpmdb/privkeys/grafana.key"}).
		WithExec([]string{"/bin/sh", "-c", "echo \"$GPG_PASSPHRASE\" | base64 -d > /root/.rpmdb/passkeys/grafana.key"}).
		WithExec([]string{"/bin/sh", "-c", "echo \"$GPG_PUBLIC_KEY\" | base64 -d > /root/.rpmdb/pubkeys/grafana.key"}).
		WithNewFile("/root/.rpmmacros", dagger.ContainerWithNewFileOpts{
			Permissions: 0400,
			Contents:    RPMMacros,
		}).
		WithExec([]string{"gpg", "--batch", "--yes", "--no-tty", "--allow-secret-key-import", "--import", "/root/.rpmdb/privkeys/grafana.key"})
}
