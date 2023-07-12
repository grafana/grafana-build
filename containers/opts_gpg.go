package containers

import (
	"github.com/grafana/grafana-build/cliutil"
)

type GPGOpts struct {
	Sign                bool
	GPGPrivateKeyBase64 string
	GPGPublicKeyBase64  string
	GPGPassphraseBase64 string
}

func GPGOptsFromFlags(c cliutil.CLIContext) *GPGOpts {
	return &GPGOpts{
		Sign:                c.Bool("sign"),
		GPGPrivateKeyBase64: c.String("gpg-private-key-base64"),
		GPGPublicKeyBase64:  c.String("gpg-public-key-base64"),
		GPGPassphraseBase64: c.String("gpg-passphrase-base64"),
	}
}
