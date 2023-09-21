package containers

import (
	"github.com/grafana/grafana-build/cliutil"
)

type PackagePublishOpts struct {
	Packages                []string
	Destination             string
	ServiceAccountKey       string
	ServiceAccountKeyBase64 string
	AccessKeyId             string
	SecretAccessKey         string
	GPGPublicKeyBase64      string
	GPGPrivateKeyBase64     string
	GPGPassphrase           string
	RemovePackages          bool
	ReplaceExisting         bool
}

func PackagePublishOptsFromFlags(c cliutil.CLIContext) *PackagePublishOpts {
	return &PackagePublishOpts{
		Packages:                c.StringSlice("package"),
		Destination:             c.String("destination"),
		ServiceAccountKey:       c.String("gcp-service-account-key"),
		ServiceAccountKeyBase64: c.String("gcp-service-account-key-base64"),
		AccessKeyId:             c.String("access-key-id"),
		SecretAccessKey:         c.String("secret-access-key"),
		GPGPublicKeyBase64:      c.String("gpg-public-key-base64"),
		GPGPrivateKeyBase64:     c.String("gpg-private-key-base64"),
		GPGPassphrase:           c.String("gpg-passphrase"),
		RemovePackages:          c.Bool("deb-remove-packages"),
		ReplaceExisting:         c.Bool("replace-existing"),
	}
}
