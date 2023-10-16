package main

import (
	"github.com/grafana/grafana-build/artifacts"
	"github.com/grafana/grafana-build/pipeline"
)

var Artifacts = []pipeline.Artifact{
	artifacts.Backend,
	artifacts.Frontend,
	artifacts.Tarball,
}
