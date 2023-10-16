package artifacts

import (
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/grafana-build/pipeline"
)

var (
	ErrorArtifactCollision = errors.New("artifact argument specifies two different artifacts")
	ErrorDuplicateArgument = errors.New("artifact argument specifies duplicate or incompatible arguments")
	ErrorNoArgument        = errors.New("could not find compatible argument for argument string")

	ErrorFlagNotFound = errors.New("no option available for the given flag")
)

func findArtifact(val string, artifacts []pipeline.Artifact) (pipeline.Artifact, error) {
	c := strings.Split(val, ":")

	// Find the artifact that is requested by `val`.
	// The artifact can be defined anywhere in the artifact string. Example: `linux/amd64:grafana:targz` or `linux/amd64:grafana:targz` are the same, where targz is the artifact.
	for _, v := range c {
		for _, a := range artifacts {
			if a.Name == v {
				return a, nil
			}
		}
	}
	return pipeline.Artifact{}, ErrorNoArgument
}

func findFlag(f []pipeline.Flag, name string) (pipeline.Flag, error) {
	for _, v := range f {
		if v.Name == name {
			return v, nil
		}
	}

	return pipeline.Flag{}, ErrorFlagNotFound
}

// Parse parses the artifact string `val` and finds the matching artifact in `artifacts`,
// populated with the options specified in the string.
func Parse(val string, artifacts []pipeline.Artifact) (pipeline.Artifact, error) {
	artifact, err := findArtifact(val, artifacts)
	if err != nil {
		return pipeline.Artifact{}, err
	}

	c := strings.Split(val, ":")
	// Given all of the other flags that were supplied for the agument string, apply them on the artifact,
	// using the artifact's own list of flags.
	for _, v := range c {
		if v == artifact.Name {
			continue
		}

		// Ensure that this argument flag is not another artifact
		if _, err := findArtifact(v, artifacts); err == nil {
			return pipeline.Artifact{}, fmt.Errorf("%w: %s", ErrorArtifactCollision, v)
		}

		f, err := findFlag(artifact.Flags, v)
		if err != nil {
			return pipeline.Artifact{}, fmt.Errorf("%w: %s", err, v)
		}

		artifact.Apply(f)
	}

	return artifact, nil
}
