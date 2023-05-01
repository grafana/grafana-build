package containers

import (
	"dagger.io/dagger"
)

func BackendTestShort(d *dagger.Client, platform dagger.Platform, dir *dagger.Directory) *dagger.Container {
	return GrafanaContainer(d, platform, GoImage, dir).
		WithExec([]string{"go", "test", "-tags", "requires_buildifer", "-short", "-covermode", "atomic", "-timeout", "5m", "./pkg/..."})
}

func BackendTestIntegration(d *dagger.Client, platform dagger.Platform, dir *dagger.Directory) *dagger.Container {
	return GrafanaContainer(d, platform, GoImage, dir).
		WithExec([]string{"go", "test", "-run", "Integration", "-covermode", "atomic", "-timeout", "5m", "./pkg/..."})
}
