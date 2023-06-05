package pipelines

import (
	"context"
	"fmt"
	"log"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/executil"
)

// ArtifactConstraint is a function that can be attached to an artifact
// definition in order to check that the artifact is only built for valid
// distributions or other setups.
type ArtifactConstraint func(ArtifactGeneratorOptions) (bool, error)

type ArtifactDefinition struct {
	Name string

	// Constraint allows you to restrict that artifact generation for certain
	// PipelineArguments. If the constraint returns false, then the artifact
	// won't be built.
	Constraint ArtifactConstraint

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
		Constraint:   ad.Constraint,
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
func (ad ArtifactDefinition) WithConstraint(con ArtifactConstraint) ArtifactDefinition {
	out := ad.clone()
	out.Constraint = con
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

type RequestedArtifact struct {
	Name    string
	Options ArtifactGeneratorOptions
}

func (ra RequestedArtifact) String() string {
	return fmt.Sprintf("%s (%s)", ra.Name, ra.Options.Distribution)
}

func GeneratateFinalArtifactList(ctx context.Context, reg *ArtifactDefinitionRegistry, artifactNames []string, args PipelineArgs) ([]RequestedArtifact, error) {
	finalArtifacts := []RequestedArtifact{}

	// Go through all the artifact trees and check if they can even be
	// built. The outcome of this is a list of all the final artifacts that
	// can eventually be exported to the user.
	for _, distro := range args.PackageOpts.Distros {
		genOpts := ArtifactGeneratorOptions{
			PipelineArgs: args,
			Distribution: distro,
		}
		for _, artifact := range artifactNames {
			req := RequestedArtifact{
				Name:    artifact,
				Options: genOpts,
			}
			checkResult, err := CheckArtifactChainConstraint(ctx, req, reg)
			if err != nil {
				return nil, err
			}
			if checkResult {
				finalArtifacts = append(finalArtifacts, req)
			}
		}
	}

	return finalArtifacts, nil
}

// CheckArtifactChainConstraint recursively checks that an artifact and all its
// dependencies can be built according to the constraints.
func CheckArtifactChainConstraint(ctx context.Context, req RequestedArtifact, reg *ArtifactDefinitionRegistry) (bool, error) {
	artifact := req.Name
	art, validArtifact := reg.Get(artifact)
	if !validArtifact {
		return false, fmt.Errorf("invalid artifact: %s", artifact)
	}

	if art.Constraint != nil {
		ok, err := art.Constraint(req.Options)
		if err != nil {
			log.Printf("Generating %s is not supported", req.String())
			return false, err
		}
		if !ok {
			log.Printf("Generating %s is not supported", req.String())
			return false, nil
		}
	}
	// Now also check recursively
	for _, requirement := range art.Requirements {
		subReq := RequestedArtifact{
			Name:    requirement,
			Options: req.Options,
		}
		reqOK, err := CheckArtifactChainConstraint(ctx, subReq, reg)
		if err != nil {
			log.Printf("Generating %s is not supported", req.String())
			return false, err
		}
		if !reqOK {
			log.Printf("Generating %s is not supported", req.String())
			return false, nil
		}
	}
	log.Printf("Generating %s is supported", req.String())
	return true, nil
}

func ConstraintLinuxOnly(opts ArtifactGeneratorOptions) (bool, error) {
	if !executil.IsLinux(opts.Distribution) {
		return false, nil
	}
	return true, nil
}
func ConstraintWindowsOnly(opts ArtifactGeneratorOptions) (bool, error) {
	if !executil.IsWindows(opts.Distribution) {
		return false, nil
	}
	return true, nil
}

func init() {
	DefaultArtifacts.Register("backend", NewArtifactDefinition().
		WithGenerator(GenerateBackendDirectory))
	DefaultArtifacts.Register("tarball", NewArtifactDefinition().
		WithGenerator(GenerateTarballDirectory).
		WithRequirement("/src/grafana/bin", "backend"))
	DefaultArtifacts.Register("docker", NewArtifactDefinition().
		WithConstraint(ConstraintLinuxOnly).
		WithRequirement("/src/tarball", "tarball"))
	DefaultArtifacts.Register("deb", NewArtifactDefinition().
		WithRequirement("/mnt/tarball", "tarball").
		WithConstraint(ConstraintLinuxOnly).
		WithGenerator(GenerateDebArtifact))
	DefaultArtifacts.Register("rpm", NewArtifactDefinition().
		WithConstraint(ConstraintLinuxOnly).
		WithGenerator(GenerateRPMArtifact).
		WithRequirement("/mnt/tarball", "tarball"))
	DefaultArtifacts.Register("windowsinstaller", NewArtifactDefinition().
		WithRequirement("/mnt/tarball", "tarball").
		WithConstraint(ConstraintWindowsOnly))
}
