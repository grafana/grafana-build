package containers

//
// import (
// 	"os"
// 	"path/filepath"
//
// 	"dagger.io/dagger"
// )
//
// // GCPAuthenticator injects authentication information into the provided container.
// type GCPAuthenticator interface {
// 	Authenticate(*dagger.Container) *dagger.Container
// }
//
// // GCPAccessToken satisfies GCPAuthenticator and injects the provided AccessToken into the filesystem and sets "GOOGLE_APPLICATION_CREDENTIALS".
// type GCPAccessToken struct {
// 	AccessToken string
// }
//
// func (a *GCPAccessToken) Authenticate(c *dagger.Container) *dagger.Container {
// 	return nil
// }
//
// // InheritedAccessToken uses `gcloud` command in the current shell to get the GCS credentials.
// // This type should really only be used when running locally.
// type GCPInheritedAuth struct{}
//
// func (a *GCPInheritedAuth) Authenticate(d *dagger.Client, c *dagger.Container) (*dagger.Container, error) {
// 	if val, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); ok {
// 		return c.WithMountedDirectory("/auth/credentials.json", d.Host().Directory(val)).WithEnvVariable("GOOGLE_APPLICATION_CREDENTIALS", "/auth/credentials.json"), nil
// 	}
//
// 	cfg, err := os.UserConfigDir()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return c.
// 		WithMountedDirectory("/auth/credentials.json", filepath.Join(cfg, "gcloud/application_default_credentials.json")).
// 		WithEnvVariable("GOOGLE_APPLICATION_CREDENTIALS", "/auth/credentials.json"), nil
// }
//
// func GCSUploadDirectory(d *dagger.Client, image string, auth GCPAuthenticator, dir *dagger.Directory, dst string) (*dagger.Container, error) {
// 	container := d.Container().From(image).
// 		WithMountedDirectory("/src", dir)
//
// 	var err error
// 	container, err = auth.Authenticate(container)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return container.WithExec([]string{"gsutil", "-m", "rsync", "-r", "/src", dst})
// }
