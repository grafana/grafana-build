package containers

import (
	"strings"

	"github.com/grafana/grafana-build/cliutil"
)

type GPGOpts struct {
	Sign                bool
	GPGPrivateKeyBase64 string
	GPGPublicKeyBase64  string
	GPGPassphrase       string
}

func GPGOptsFromFlags(c cliutil.CLIContext) *GPGOpts {
	return &GPGOpts{
		Sign:                c.Bool("sign"),
		GPGPrivateKeyBase64: strings.ReplaceAll(c.String("gpg-private-key-base64"), "\n", ""),
		GPGPublicKeyBase64:  strings.ReplaceAll(c.String("gpg-public-key-base64"), "\n", ""),
		GPGPassphrase:       c.String("gpg-passphrase"),
	}
}
