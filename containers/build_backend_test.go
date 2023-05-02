package containers_test

import (
	"fmt"
	"testing"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"github.com/grafana/grafana-build/executil"
)

func ExpectPlatform(t *testing.T, expect, result dagger.Platform) {
	t.Helper()

	if expect != result {
		t.Errorf("Unexpected platform '%s', expected '%s'", expect, result)
	}
}

func TestBuilderPlatform(t *testing.T) {
	var (
		linuxAMD64 = dagger.Platform("linux/amd64")
		linuxARM64 = dagger.Platform("linux/amd64")
	)

	tests := []struct {
		InputPlatform    dagger.Platform
		Distro           executil.Distribution
		ExpectedPlatform dagger.Platform
	}{
		{
			InputPlatform:    linuxAMD64,
			Distro:           executil.DistLinuxAMD64,
			ExpectedPlatform: linuxAMD64,
		},
		{
			InputPlatform:    linuxARM64,
			Distro:           executil.DistLinuxARM64,
			ExpectedPlatform: linuxARM64,
		},
	}

	for i, v := range tests {
		title := fmt.Sprintf("[%d/%d] distro: '%s', platform: '%s' should use platform '%s'", i+1, len(tests), v.Distro, v.InputPlatform, v.ExpectedPlatform)
		t.Run(title, func(t *testing.T) {
			p := containers.BuilderPlatform(v.Distro, v.InputPlatform)
			ExpectPlatform(t, v.ExpectedPlatform, p)
		})
	}
}
