package containers

import (
	"dagger.io/dagger"
)

func WithDarwinAMD64Toolchain(c *dagger.Container) *dagger.Container {
	// Download osxcross...
	// set CC environment variable
	return c
}

func WithDarwinARM64Toolchain(c *dagger.Container) *dagger.Container {
	return c
}
