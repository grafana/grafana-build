package containers

import (
	"encoding/base64"

	"github.com/grafana/grafana-build/cliutil"
)

type GPGOpts struct {
	Sign          bool
	GPGPrivateKey string
	GPGPublicKey  string
	GPGPassphrase string
}

func GPGOptsFromFlags(c cliutil.CLIContext) (*GPGOpts, error) {
	gpgPrivateKey, err := base64.StdEncoding.DecodeString(c.String("gpg-private-key"))
	if err != nil {
		return nil, err
	}
	gpgPublicKey, err := base64.StdEncoding.DecodeString(c.String("gpg-public-key"))
	if err != nil {
		return nil, err
	}
	gpgPassphrase, err := base64.StdEncoding.DecodeString(c.String("gpg-passphrase"))
	if err != nil {
		return nil, err
	}
	return &GPGOpts{
		Sign:          c.Bool("sign"),
		GPGPrivateKey: string(gpgPrivateKey),
		GPGPublicKey:  string(gpgPublicKey),
		GPGPassphrase: string(gpgPassphrase),
	}, nil
}
