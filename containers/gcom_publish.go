package containers

import (
	"context"
	"encoding/json"
	"fmt"

	"dagger.io/dagger"
)

type GCOMVersionPayload struct {
	Version         string `json:"version"`         // "10.0.3"
	ReleaseDate     string `json:"releaseDate"`     // "2023-07-26T08:20:16.628278891Z"
	Stable          bool   `json:"stable"`          // true
	Beta            bool   `json:"beta"`            // false
	Nightly         bool   `json:"nightly"`         // false
	WhatsNewURL     string `json:"whatsNewUrl"`     // "https://grafana.com/docs/grafana/next/whatsnew/whats-new-in-v10-0/"
	ReleaseNotesURL string `json:"releaseNotesUrl"` // "https://grafana.com/docs/grafana/next/release-notes/"
}

type GCOMPackagePayload struct {
	OS     string `json:"os"`     // "deb"
	URL    string `json:"url"`    // "https://dl.grafana.com/oss/release/grafana_10.0.3_arm64.deb"
	Sha256 string `json:"sha256"` // "78a718816dd556198cfa3007dd594aaf1d80886decae8c4bd0f615bd3f118279\n"
	Arch   string `json:"arch"`   // "arm64"
}

// PublishGCOM publishes a package to grafana.com.
func PublishGCOM(ctx context.Context, d *dagger.Client, versionPayload *GCOMVersionPayload, packagePayload *GCOMPackagePayload, opts *GCOMOpts) error {
	versionApiUrl := fmt.Sprintf("%s/api/grafana/versions", opts.GCOMUrl)
	packagesApiUrl := fmt.Sprintf("%s/api/grafana/versions/%s/packages", opts.GCOMUrl, versionPayload.Version)

	jsonVersionPayload, err := json.Marshal(versionPayload)
	if err != nil {
		return err
	}

	jsonPackagePayload, err := json.Marshal(packagePayload)
	if err != nil {
		return err
	}

	apiKeySecret := d.SetSecret("gcom-api-key", opts.GCOMApiKey)

	_, err = d.Container().From("alpine").
		WithSecretVariable("GCOM_API_KEY", apiKeySecret).
		WithExec([]string{"apk", "add", "curl"}).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf(`curl -H "Content-Type: application/json" -H "Authorization: Bearer $GCOM_API_KEY" -d '%s' %s`, string(jsonVersionPayload), versionApiUrl)}).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf(`curl -H "Content-Type: application/json" -H "Authorization: Bearer $GCOM_API_KEY" -d '%s' %s`, string(jsonPackagePayload), packagesApiUrl)}).
		Sync(ctx)

	if err != nil {
		return err
	}

	return nil
}
