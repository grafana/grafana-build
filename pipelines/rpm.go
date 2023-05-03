package pipelines

import (
	"context"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

// RPM uses the grafana package given by the '--package' argument and creates a .rpm installer.
// It accepts publish args, so you can place the file in a local or remote destination.
func RPM(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	return PackageInstaller(ctx, d, args, InstallerOpts{
		PackageType: "rpm",
		ConfigFiles: [][]string{
			{"/src/packaging/rpm/sysconfig/grafana-server", "/pkg/etc/sysconfig/grafana-server"},
			{"/src/packaging/rpm/init.d/grafana-server", "/pkg/etc/init.d/grafana-server"},
			{"/src/packaging/rpm/systemd/grafana-server.service", "/pkg/usr/lib/systemd/system/grafana-server.service"},
		},
		AfterInstall: "/src/packaging/rpm/control/postinst",
		Depends: []string{
			"/sbin/service",
			"chkconfig",
			"fontconfig",
			"freetype",
			"urw-fonts",
		},
		ExtraArgs: []string{
			"--rpm-posttrans=/src/packaging/rpm/control/posttrans",
			"--rpm-digest=sha256",
		},
		EnvFolder: "/pkg/etc/sysconfig",
		RPMSign:   args.SignOpts.Sign,
		Container: RPMContainer(d, args),
	})
}

func RPMContainer(d *dagger.Client, args PipelineArgs) *dagger.Container {
	container := containers.FPMContainer(d).
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-yq", "rpm"})
	if !args.SignOpts.Sign {
		return container
	}
	return container.
		WithNewFile("/root/.rpmmacros", dagger.ContainerWithNewFileOpts{
			Contents: `%_signature gpg
%_gpg_path /root/.gnupg
%_gpg_name "Grafana"
%_gpgbin /usr/bin/gpg
%__gpg_sign_cmd %{__gpg} gpg --batch --yes --pinentry-mode loopback --no-armor --passphrase-file /root/.rpmdb/passkeys/grafana.key --no-secmem-warning -u "%{_gpg_name}" -sbo %{__signature_filename} %{__plaintext_filename}`,
		}).
		WithExec([]string{"cat", "/root/.rpmmacros"}).
		WithNewFile("/root/.rpmdb/privkeys/grafana.key", dagger.ContainerWithNewFileOpts{
			Contents: args.SignOpts.GPGPrivateKey,
		}).
		WithNewFile("/root/.rpmdb/pubkeys/grafana.key", dagger.ContainerWithNewFileOpts{
			Contents: args.SignOpts.GPGPublicKey,
		}).
		WithNewFile("/root/.rpmdb/passkeys/grafana.key", dagger.ContainerWithNewFileOpts{
			Contents: args.SignOpts.GPGPassphrase,
		}).
		WithExec([]string{"gpg", "--batch", "--yes", "--no-tty", "--allow-secret-key-import", "--import", "/root/.rpmdb/privkeys/grafana.key"})
}
