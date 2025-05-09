package main

import (
	"github.com/grafana/grafana-build/artifacts"
)

var Artifacts = map[string]artifacts.Initializer{
	"backend":           artifacts.BackendInitializer,
	"frontend":          artifacts.FrontendInitializer,
	"npm":               artifacts.NPMPackagesInitializer,
	"targz":             artifacts.TargzInitializer,
	"zip":               artifacts.ZipInitializer,
	"deb":               artifacts.DebInitializer,
	"rpm":               artifacts.RPMInitializer,
	"docker":            artifacts.DockerInitializer,
	"docker-pro":        artifacts.ProDockerInitializer,
	"docker-enterprise": artifacts.EntDockerInitializer,
	"storybook":         artifacts.StorybookInitializer,
	"msi":               artifacts.MSIInitializer,
	"version":           artifacts.VersionInitializer,
}
