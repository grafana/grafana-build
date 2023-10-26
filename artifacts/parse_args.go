package artifacts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/grafana/grafana-build/pipeline"
)

var (
	ErrorArtifactCollision = errors.New("artifact argument specifies two different artifacts")
	ErrorDuplicateArgument = errors.New("artifact argument specifies duplicate or incompatible arguments")
	ErrorNoArtifact        = errors.New("could not find compatible artifact for argument string")

	ErrorFlagNotFound = errors.New("no option available for the given flag")
)

func findInitializer(val string, initializers map[string]Initializer) (Initializer, error) {
	c := strings.Split(val, ":")
	var initializer *Initializer

	// Find the artifact that is requested by `val`.
	// The artifact can be defined anywhere in the artifact string. Example: `linux/amd64:grafana:targz` or `linux/amd64:grafana:targz` are the same, where targz is the artifact.
	for _, v := range c {
		n, ok := initializers[v]
		if !ok {
			continue
		}
		if initializer != nil {
			return Initializer{}, fmt.Errorf("%s: %w", val, ErrorArtifactCollision)
		}

		initializer = &n
	}

	if initializer == nil {
		return Initializer{}, fmt.Errorf("%s: %w", val, ErrorNoArtifact)
	}

	return *initializer, nil
}

func findFlag(f []pipeline.Flag, name string) (pipeline.Flag, error) {
	for _, v := range f {
		if v.Name == name {
			return v, nil
		}
	}

	return pipeline.Flag{}, ErrorFlagNotFound
}

// The ArtifactsFromStrings function should provide all of the necessary arguments to produce each artifact
// dleimited by colons. It's a repeated flag, so all permutations are stored in 1 instance of the ArtifactsFlag struct.
// Examples:
// * targz:linux/amd64 -- Will produce a "Grafana" tar.gz for "linux/amd64".
// * targz:enterprise:linux/amd64 -- Will produce a "Grafana" tar.gz for "linux/amd64".
func ArtifactsFromStrings(ctx context.Context, log *slog.Logger, a []string, registered map[string]Initializer, state pipeline.StateHandler) ([]*pipeline.Artifact, error) {
	artifacts := make([]*pipeline.Artifact, len(a))
	for i, v := range a {
		n, err := Parse(ctx, log, v, registered, state)
		if err != nil {
			return nil, err
		}

		artifacts[i] = n
	}

	return artifacts, nil
}

// Parse parses the artifact string `val` and finds the matching initializer.
func Parse(ctx context.Context, log *slog.Logger, val string, initializers map[string]Initializer, state pipeline.StateHandler) (*pipeline.Artifact, error) {
	initializer, err := findInitializer(val, initializers)
	if err != nil {
		return nil, err
	}

	initializerFunc := initializer.InitializerFunc
	// TODO soon, the initializer might need more info about flags
	return initializerFunc(ctx, log, val, state)
}
