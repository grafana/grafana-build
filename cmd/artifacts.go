package main

import (
	"github.com/grafana/grafana-build/artifacts"
)

var Artifacts = map[string]artifacts.Initializer{
	// artifacts.BackendKey: artifacts.BackendInitializer,
	"frontend": artifacts.FrontendInitializer,
	"targz":    artifacts.TargzInitializer,
	"deb":      artifacts.DebInitializer,
	"rpm":      artifacts.RPMInitializer,
	"docker":   artifacts.DockerInitializer,
}
