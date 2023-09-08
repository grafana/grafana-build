package main

import (
	"github.com/grafana/grafana-build/cmd/artifacts"
	"github.com/grafana/grafana-build/pipeline"
)

var Artifacts = []*pipeline.Artifact{
	artifacts.Backend,
	artifacts.Frontend,
	artifacts.Tarball,
}
