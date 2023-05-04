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

func RPMContainer(d *dagger.Client, opts *SignOpts) *dagger.Container {
	container := FPMContainer(d).
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-yq", "rpm"})
	if !opts.Sign {
		return container
	}
	return container.
		WithExec([]string{"apt-get", "install", "-yq", "gnupg2"}).
		WithNewFile("/root/.rpmmacros", dagger.ContainerWithNewFileOpts{
			Permissions: 0400,
			Contents:    RPMMacros,
		}).
		WithNewFile("/root/.rpmdb/privkeys/grafana.key", dagger.ContainerWithNewFileOpts{
			Permissions: 0400,
			Contents:    opts.GPGPrivateKey,
		}).
		WithNewFile("/root/.rpmdb/pubkeys/grafana.key", dagger.ContainerWithNewFileOpts{
			Permissions: 0400,
			Contents:    opts.GPGPublicKey,
		}).
		WithNewFile("/root/.rpmdb/passkeys/grafana.key", dagger.ContainerWithNewFileOpts{
			Permissions: 0400,
			Contents:    opts.GPGPassphrase,
		}).
		WithExec([]string{"gpg", "--batch", "--yes", "--no-tty", "--allow-secret-key-import", "--import", "/root/.rpmdb/privkeys/grafana.key"})
	// WithExec([]string{"rpm", "--import", "/root/.rpmdb/pubkeys/grafana.key"})
}
