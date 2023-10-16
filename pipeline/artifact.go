package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"dagger.io/dagger"
)

var (
	ErrorNotADirectory      = errors.New("not a directory argument")
	ErrorOptionNotSet       = errors.New("expected option not set")
	ErrorDependencyNotFound = errors.New("dependency not found")
)

type ArtifactType int

const (
	ArtifactTypeFile ArtifactType = iota
	ArtifactTypeDirectory
)

type ArtifactContainerOpts struct {
	Log      *slog.Logger
	Client   *dagger.Client
	Platform dagger.Platform
	State    StateHandler
}

type ArtifactBuildOpts struct {
	ContainerOpts *ArtifactContainerOpts
	// Dependencies are artifacts that this artifact depends on.
	Dependencies map[string]Artifact
	Builder      *dagger.Container
}

func (o *ArtifactBuildOpts) Dependency(artifact Artifact) (Artifact, error) {
	if o.Dependencies == nil {
		o.Dependencies = map[string]Artifact{}
	}

	v, ok := o.Dependencies[artifact.Name]
	if !ok {
		return Artifact{}, fmt.Errorf("%s: %w", artifact.Name, ErrorDependencyNotFound)
	}

	return v, nil
}

type ArtifactPublishFileOpts struct {
	Log    *slog.Logger
	Client *dagger.Client
	State  StateHandler

	File *dagger.File
}

type ArtifactPublishDirOpts struct {
	Log    *slog.Logger
	Client *dagger.Client
	State  StateHandler

	Directory *dagger.Directory
}

type (
	ContainerFunc func(ctx context.Context, opts *ArtifactContainerOpts) (*dagger.Container, error)

	BuildFileFunc func(ctx context.Context, opts *ArtifactBuildOpts) (*dagger.File, error)
	BuildDirFunc  func(ctx context.Context, opts *ArtifactBuildOpts) (*dagger.Directory, error)

	PublishFileFunc func(ctx context.Context, opts *ArtifactPublishFileOpts) error
	PublishDirFunc  func(ctx context.Context, opts *ArtifactPublishDirOpts) error

	FileNameFunc func(ctx context.Context, a Artifact, state StateHandler) (string, error)
)

type Artifact struct {
	Name      string
	Type      ArtifactType
	Requires  []Artifact
	Arguments []Argument
	Flags     []Flag
	options   map[string]any

	Builder   ContainerFunc
	Publisher ContainerFunc

	BuildFileFunc BuildFileFunc
	BuildDirFunc  BuildDirFunc

	PublishFileFunc PublishFileFunc
	PublishDirFunc  PublishDirFunc

	FileNameFunc FileNameFunc
}

func (a *Artifact) Apply(f Flag) {
	for k, v := range f.Options {
		a.SetOption(k, v)
	}
}

func (a *Artifact) SetOption(key, value string) {
	if a.options == nil {
		a.options = map[string]any{}
	}

	a.options[key] = value
}

func (a *Artifact) Option(key string) (string, error) {
	if a.options == nil {
		return "", fmt.Errorf("%w: %s", ErrorOptionNotSet, key)
	}

	v, ok := a.options[key]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrorOptionNotSet, key)
	}

	return v.(string), nil
}

func (a *Artifact) Directory(ctx context.Context, opts *ArtifactContainerOpts) (*dagger.Directory, error) {
	if a.Type != ArtifactTypeDirectory {
		return nil, fmt.Errorf("%s: %w", a.Name, ErrorNotADirectory)
	}

	builder, err := a.Builder(ctx, opts)

	if err != nil {
		return nil, err
	}

	return a.BuildDirFunc(ctx, &ArtifactBuildOpts{
		Builder: builder,
	})
}

func (a *Artifact) File(ctx context.Context) (*dagger.File, error) {
	return nil, nil
}
