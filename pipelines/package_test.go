package pipelines_test

import (
	"testing"

	"github.com/grafana/grafana-build/executil"
	"github.com/grafana/grafana-build/pipelines"
)

func TestTarFilename(t *testing.T) {
	t.Run("It should use the correct name if Enterprise is false", func(t *testing.T) {
		distro := executil.Distribution("plan9/amd64")
		args := pipelines.PipelineArgs{
			Version:         "v1.0.1-test",
			BuildID:         "333",
			BuildEnterprise: false,
		}

		expected := "grafana_v1.0.1-test_plan9_amd64_333.tar.gz"
		if name := pipelines.TarFilename(args.Version, args.BuildID, args.BuildEnterprise, distro); name != expected {
			t.Errorf("name '%s' does not match expected name '%s'", name, expected)
		}
	})
	t.Run("It should use the correct name if Enterprise is true", func(t *testing.T) {
		distro := executil.Distribution("plan9/amd64")
		args := pipelines.PipelineArgs{
			Version:         "v1.0.1-test",
			BuildID:         "333",
			BuildEnterprise: true,
		}

		expected := "grafana-enterprise_v1.0.1-test_plan9_amd64_333.tar.gz"
		if name := pipelines.TarFilename(args.Version, args.BuildID, args.BuildEnterprise, distro); name != expected {
			t.Errorf("name '%s' does not match expected name '%s'", name, expected)
		}
	})
	t.Run("It should use include the arch version if one is supplied in the distro", func(t *testing.T) {
		distro := executil.Distribution("plan9/arm/v6")
		args := pipelines.PipelineArgs{
			Version:         "v1.0.1-test",
			BuildID:         "333",
			BuildEnterprise: true,
		}

		expected := "grafana-enterprise_v1.0.1-test_plan9_arm-6_333.tar.gz"
		if name := pipelines.TarFilename(args.Version, args.BuildID, args.BuildEnterprise, distro); name != expected {
			t.Errorf("name '%s' does not match expected name '%s'", name, expected)
		}
	})
}
