package containers

import (
	"context"
	"encoding/json"
	"fmt"

	"dagger.io/dagger"
)

type GCOMVersionPayload struct {
	Version         string `json:"version"`
	ReleaseDate     string `json:"releaseDate"`
	Stable          bool   `json:"stable"`
	Beta            bool   `json:"beta"`
	Nightly         bool   `json:"nightly"`
	WhatsNewURL     string `json:"whatsNewUrl"`
	ReleaseNotesURL string `json:"releaseNotesUrl"`
}

type GCOMPackagePayload struct {
	OS     string `json:"os"`
	URL    string `json:"url"`
	Sha256 string `json:"sha256"`
	Arch   string `json:"arch"`
}

// PublishGCOMVersion publishes a version to grafana.com.
func PublishGCOMVersion(ctx context.Context, d *dagger.Client, versionPayload *GCOMVersionPayload, opts *GCOMOpts) error {
	versionApiUrl := fmt.Sprintf("%s/api/grafana/versions", opts.URL)

	jsonVersionPayload, err := json.Marshal(versionPayload)
	if err != nil {
		return err
	}

	apiKeySecret := d.SetSecret("gcom-api-key", opts.ApiKey)

	_, err = d.Container().From("alpine/curl").
		WithSecretVariable("GCOM_API_KEY", apiKeySecret).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf(`curl -H "Content-Type: application/json" -H "Authorization: Bearer $GCOM_API_KEY" -d '%s' %s`, string(jsonVersionPayload), versionApiUrl)}).
		Sync(ctx)

	if err != nil {
		return err
	}

	return nil
}

// PublishGCOMPackage publishes a package to grafana.com.
func PublishGCOMPackage(ctx context.Context, d *dagger.Client, packagePayload *GCOMPackagePayload, opts *GCOMOpts, version string) error {
	packagesApiUrl := fmt.Sprintf("%s/api/grafana/versions/%s/packages", opts.URL, version)

	jsonPackagePayload, err := json.Marshal(packagePayload)
	if err != nil {
		return err
	}

	apiKeySecret := d.SetSecret("gcom-api-key", opts.ApiKey)

	_, err = d.Container().From("alpine/curl").
		WithSecretVariable("GCOM_API_KEY", apiKeySecret).
		WithExec([]string{"/bin/sh", "-c", fmt.Sprintf(`curl -H "Content-Type: application/json" -H "Authorization: Bearer $GCOM_API_KEY" -d '%s' %s`, string(jsonPackagePayload), packagesApiUrl)}).
		Sync(ctx)

	if err != nil {
		return err
	}

	return nil
}
