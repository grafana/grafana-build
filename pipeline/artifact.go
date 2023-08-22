package pipeline

import (
	"context"

	"dagger.io/dagger"
)

type (
	ContainerFunc func(d *dagger.Client, platform dagger.Platform, args []Argument) *dagger.Container
	BuildFunc     func(ctx context.Context, d *dagger.Client, c *dagger.Container, args []Argument) (*dagger.File, error)
	PublishFunc   func(ctx context.Context, d *dagger.Client, c *dagger.Container, args []Argument) error
)

type Artifact struct {
	Name      string
	Requires  []*Artifact
	Arguments []Argument

	Builder     ContainerFunc
	BuildFunc   BuildFunc
	Publisher   ContainerFunc
	PublishFunc PublishFunc
}
