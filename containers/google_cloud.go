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
	JSONFile string
}

func (a *GCPServiceAccount) Authenticate(d *dagger.Client, c *dagger.Container) (*dagger.Container, error) {
	return c.
		WithMountedFile(
			"/opt/service_account.json",
			d.Host().Directory(filepath.Dir(a.JSONFile)).File(filepath.Base(a.JSONFile)),
		).
		WithExec([]string{"gcloud", "auth", "activate-service-account", "--key-file", "/opt/service_account.json"}), nil
}

func NewGCPServiceAccount(filepath string) *GCPServiceAccount {
	return &GCPServiceAccount{
		JSONFile: filepath,
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

	return container.WithExec([]string{"gsutil", "-m", "rsync", "-r", "/src", dst}), nil
}

func GCSUploadFile(d *dagger.Client, image string, auth GCPAuthenticator, file *dagger.File, dst string) (*dagger.Container, error) {
	container := d.Container().From(image).
		WithMountedFile("/src/file", file)

	var err error
	container, err = auth.Authenticate(d, container)
	if err != nil {
		return nil, err
	}

	return container.WithExec([]string{"gsutil", "cp", "/src/file", dst}), nil
}
