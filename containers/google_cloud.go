package containers

import (
	"os"
	"path/filepath"

	"dagger.io/dagger"
)

const GoogleCloudImage = "google/cloud-sdk:alpine"

// GCPAuthenticator injects authentication information into the provided container.
type GCPAuthenticator interface {
	Authenticate(*dagger.Client, *dagger.Container) (*dagger.Container, error)
}

// GCPServiceAccount satisfies GCPAuthenticator and injects the provided ServiceAccount into the filesystem and adds a 'gcloud auth activate-service-account'
type GCPServiceAccount struct {
	DaggerFile *dagger.File
	JSONFile   string
}

func (a *GCPServiceAccount) Authenticate(d *dagger.Client, c *dagger.Container) (*dagger.Container, error) {
	container := c.WithMountedFile(
		"/opt/service_account.json",
		d.Host().Directory(filepath.Dir(a.JSONFile)).File(filepath.Base(a.JSONFile)),
	)

	if a.DaggerFile != nil {
		container = container.WithMountedFile("/opt/service_account.json", a.DaggerFile)
	}

	return container.WithExec([]string{"gcloud", "auth", "activate-service-account", "--key-file", "/opt/service_account.json"}), nil
}

func NewGCPServiceAccount(filepath string) *GCPServiceAccount {
	return &GCPServiceAccount{
		JSONFile: filepath,
	}
}

func NewGCPServiceAccountWithFile(file *dagger.File) *GCPServiceAccount {
	return &GCPServiceAccount{
		DaggerFile: file,
	}
}

// InheritedServiceAccount uses `gcloud` command in the current shell to get the GCS credentials.
// This type should really only be used when running locally.
type GCPInheritedAuth struct{}

func (a *GCPInheritedAuth) Authenticate(d *dagger.Client, c *dagger.Container) (*dagger.Container, error) {
	if val, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); ok {
		return c.WithMountedDirectory("/auth/credentials.json", d.Host().Directory(val)).WithEnvVariable("GOOGLE_APPLICATION_CREDENTIALS", "/auth/credentials.json"), nil
	}

	cfg, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	return c.WithMountedDirectory("/root/.config/gcloud", d.Host().Directory(filepath.Join(cfg, "gcloud"))), nil
}

func GCSUploadDirectory(d *dagger.Client, image string, auth GCPAuthenticator, dir *dagger.Directory, dst string) (*dagger.Container, error) {
	container := d.Container().From(image).
		WithMountedDirectory("/src", dir)

	var err error
	container, err = auth.Authenticate(d, container)
	if err != nil {
		return nil, err
	}

	secret := d.SetSecret("gcs-destination", dst)
	container = container.WithSecretVariable("GCS_DESTINATION", secret)

	return container.WithExec([]string{"/bin/sh", "-c", "gsutil -m rsync -r /src ${GCS_DESTINATION}"}), nil
}

func GCSUploadFile(d *dagger.Client, image string, auth GCPAuthenticator, file *dagger.File, dst string) (*dagger.Container, error) {
	container := d.Container().From(image).
		WithMountedFile("/src/file", file)

	var err error
	container, err = auth.Authenticate(d, container)
	if err != nil {
		return nil, err
	}
	secret := d.SetSecret("gcs-destination", dst)
	container = container.WithSecretVariable("GCS_DESTINATION", secret)
	return container.WithExec([]string{"/bin/sh", "-c", "gsutil cp /src/file ${GCS_DESTINATION}"}), nil
}
