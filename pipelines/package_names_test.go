package pipelines_test

import (
	"testing"

	"github.com/grafana/grafana-build/executil"
	"github.com/grafana/grafana-build/pipelines"
)

func TestTarFilename(t *testing.T) {
	t.Run("It should use the correct name if Enterprise is false", func(t *testing.T) {
		distro := executil.Distribution("plan9/amd64")
		opts := pipelines.TarFileOpts{
			IsEnterprise: false,
			Version:      "v1.0.1-test",
			BuildID:      "333",
			Distro:       distro,
		}

		expected := "grafana_v1.0.1-test_plan9_amd64_333.tar.gz"
		if name := pipelines.TarFilename(opts); name != expected {
			t.Errorf("name '%s' does not match expected name '%s'", name, expected)
		}
	})
	t.Run("It should use the correct name if Enterprise is true", func(t *testing.T) {
		distro := executil.Distribution("plan9/amd64")
		opts := pipelines.TarFileOpts{
			IsEnterprise: true,
			Version:      "v1.0.1-test",
			BuildID:      "333",
			Distro:       distro,
		}

		expected := "grafana-enterprise_v1.0.1-test_plan9_amd64_333.tar.gz"
		if name := pipelines.TarFilename(opts); name != expected {
			t.Errorf("name '%s' does not match expected name '%s'", name, expected)
		}
	})
	t.Run("It should use include the arch version if one is supplied in the distro", func(t *testing.T) {
		distro := executil.Distribution("plan9/arm/v6")
		opts := pipelines.TarFileOpts{
			IsEnterprise: true,
			Version:      "v1.0.1-test",
			BuildID:      "333",
			Distro:       distro,
		}

		expected := "grafana-enterprise_v1.0.1-test_plan9_arm-6_333.tar.gz"
		if name := pipelines.TarFilename(opts); name != expected {
			t.Errorf("name '%s' does not match expected name '%s'", name, expected)
		}
	})
}

func TestOptsFromFile(t *testing.T) {
	t.Run("It should get the correct tar file opts from a valid name", func(t *testing.T) {
		name := "grafana-enterprise_v1.0.1-test_plan9_arm-6_333.tar.gz"
		distro := executil.Distribution("plan9/arm/v6")
		expect := pipelines.TarFileOpts{
			IsEnterprise: true,
			Version:      "v1.0.1-test",
			BuildID:      "333",
			Distro:       distro,
		}
		got := pipelines.TarOptsFromFileName(name)
		if got.IsEnterprise != expect.IsEnterprise {
			t.Errorf("got.IsEnterprise != expect.IsEnterprise, expected '%t'", expect.IsEnterprise)
		}
		if got.Version != expect.Version {
			t.Errorf("got.Version != expect.Version, expected '%s', got '%s'", expect.Version, got.Version)
		}
		if got.BuildID != expect.BuildID {
			t.Errorf("got.BuildID != expect.BuildID, expected '%s', got '%s'", expect.BuildID, got.BuildID)
		}
		if got.Distro != expect.Distro {
			t.Errorf("got.Distro != expect.Distro, expected '%s', got '%s'", expect.Distro, got.Distro)
		}
	})
}
