package containers

import (
	"github.com/grafana/grafana-build/cliutil"
)

type GPGOpts struct {
	Sign          bool
	GPGPrivateKey string
	GPGPublicKey  string
	GPGPassphrase string
}

func GPGOptsFromFlags(c cliutil.CLIContext) *GPGOpts {
	return &GPGOpts{
		Sign:          c.Bool("sign"),
		GPGPrivateKey: c.String("gpg-private-key"),
		GPGPublicKey:  c.String("gpg-public-key"),
		GPGPassphrase: c.String("gpg-passphrase"),
	}
}
