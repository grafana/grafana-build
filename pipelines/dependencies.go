package pipelines

import (
	"context"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/executil"
)

type ArtifactDefinition struct {
	Name string

	// Requirements is basically a map of mount points with their respective
	// sources. For instance, the package artifact requires frontend and
	// backend directories to be mounted in specific folder for them to be
	// combined
	Requirements map[string]string

	// Generator is responsible for taking the commandline options/pipeline
	// configuration and generating a single dagger.Directory out of it.
	Generator ArtifactGenerator
}

func NewArtifactDefinition() ArtifactDefinition {
	return ArtifactDefinition{
		Requirements: make(map[string]string),
	}
}

func (ad ArtifactDefinition) clone() ArtifactDefinition {
	requirements := make(map[string]string)
	for k, v := range ad.Requirements {
		requirements[k] = v
	}
	return ArtifactDefinition{
		Name:         ad.Name,
		Requirements: requirements,
		Generator:    ad.Generator,
	}
}

func (ad ArtifactDefinition) WithRequirement(mount, artifact string) ArtifactDefinition {
	out := ad.clone()
	out.Requirements[mount] = artifact
	return out
}

func (ad ArtifactDefinition) WithGenerator(gen ArtifactGenerator) ArtifactDefinition {
	out := ad.clone()
	out.Generator = gen
	return out
}

type ArtifactGeneratorOptions struct {
	Distribution executil.Distribution
	PipelineArgs PipelineArgs
}

type ArtifactGenerator func(ctx context.Context, d *dagger.Client, src *dagger.Directory, opts ArtifactGeneratorOptions, mounts map[string]*dagger.Directory) (*dagger.Directory, error)

func NewArtifactDefinitionRegistry() *ArtifactDefinitionRegistry {
	return &ArtifactDefinitionRegistry{
		data: make(map[string]ArtifactDefinition),
	}
}

type ArtifactDefinitionRegistry struct {
	data map[string]ArtifactDefinition
}

func (reg *ArtifactDefinitionRegistry) Register(name string, ad ArtifactDefinition) {
	ad.Name = name
	reg.data[name] = ad
}

func (reg *ArtifactDefinitionRegistry) Get(name string) (ArtifactDefinition, bool) {
	ad, ok := reg.data[name]
	return ad, ok
}

var DefaultArtifacts = NewArtifactDefinitionRegistry()

func init() {
	DefaultArtifacts.Register("backend", NewArtifactDefinition().
		WithGenerator(GenerateBackendDirectory))
	DefaultArtifacts.Register("tarball", NewArtifactDefinition().
		WithGenerator(GenerateTarballDirectory).
		WithRequirement("/src/grafana/bin", "backend"))
	DefaultArtifacts.Register("docker", NewArtifactDefinition().WithRequirement("/src/tarball", "tarball"))
	DefaultArtifacts.Register("deb", NewArtifactDefinition().
		WithRequirement("/mnt/tarball", "tarball").
		WithGenerator(GenerateDebArtifact))
	DefaultArtifacts.Register("rpm", NewArtifactDefinition().WithRequirement("/mnt/tarball", "tarball"))
	DefaultArtifacts.Register("windowsinstaller", NewArtifactDefinition().WithRequirement("/mnt/tarball", "tarball"))
}
