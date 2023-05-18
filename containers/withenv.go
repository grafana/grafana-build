package containers

import (
	"dagger.io/dagger"
)

func WithEnv(c *dagger.Container, env map[string]string) *dagger.Container {
	container := c
	for k, v := range env {
		container = container.WithEnvVariable(k, v)
	}

	return container
}
