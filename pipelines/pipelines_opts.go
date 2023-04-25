package pipelines

import "github.com/grafana/grafana-build/containers"

func GetPublishOpts(c CLIContext) *containers.PublishOpts {
	return &containers.PublishOpts{
		Destination: c.String("destination"),
		GCSOpts: &containers.GCSOpts{
			ServiceAccountKeyBase64: c.String("gcp-service-account-key-base64"),
			ServiceAccountKey:       c.String("gcp-service-account-key"),
		},
	}
}

func GetPackageInputOpts(c CLIContext) *containers.PackageInputOpts {
	return &containers.PackageInputOpts{
		Package: c.String("package"),
	}
}
