package pipelines_test

import (
	"testing"

	"github.com/grafana/grafana-build/executil"
	"github.com/grafana/grafana-build/pipelines"
)

func TestWithoutExt(t *testing.T) {
	names := map[string]string{
		"grafana_v1.0.1-test_333_plan9_amd64.tar.gz":                          "grafana_v1.0.1-test_333_plan9_amd64",
		"grafana-enterprise_v1.0.1-test_333_plan9_amd64.tar.gz":               "grafana-enterprise_v1.0.1-test_333_plan9_amd64",
		"grafana-enterprise_v1.0.1-test_333_plan9_arm-6.tar.gz":               "grafana-enterprise_v1.0.1-test_333_plan9_arm-6",
		"grafana-enterprise_v1.0.1-test_333_plan9_amd64.deb":                  "grafana-enterprise_v1.0.1-test_333_plan9_amd64",
		"grafana-enterprise_v1.0.1-test_333_plan9_arm-6.deb":                  "grafana-enterprise_v1.0.1-test_333_plan9_arm-6",
		"grafana-enterprise_v1.0.1-test_333_plan9_arm-6.docker.tar.gz":        "grafana-enterprise_v1.0.1-test_333_plan9_arm-6",
		"grafana-enterprise_v1.0.1-test_333_plan9_arm-6.ubuntu.docker.tar.gz": "grafana-enterprise_v1.0.1-test_333_plan9_arm-6",
	}

	for k, v := range names {
		if n := pipelines.WithoutExt(k); n != v {
			t.Errorf("Expected '%s' without file name to equal '%s' but got '%s'", k, v, n)
		}
	}
}

func TestDestinationName(t *testing.T) {
	names := map[string]string{
		"grafana_v1.0.1-test_333_plan9_amd64.tar.gz":                          "grafana_v1.0.1-test_333_plan9_amd64.ext",
		"grafana-enterprise_v1.0.1-test_333_plan9_amd64.tar.gz":               "grafana-enterprise_v1.0.1-test_333_plan9_amd64.ext",
		"grafana-enterprise_v1.0.1-test_333_plan9_arm-6.tar.gz":               "grafana-enterprise_v1.0.1-test_333_plan9_arm-6.ext",
		"grafana-enterprise_v1.0.1-test_333_plan9_amd64.deb":                  "grafana-enterprise_v1.0.1-test_333_plan9_amd64.ext",
		"grafana-enterprise_v1.0.1-test_333_plan9_arm-6.deb":                  "grafana-enterprise_v1.0.1-test_333_plan9_arm-6.ext",
		"grafana-enterprise_v1.0.1-test_333_plan9_arm-6.docker.tar.gz":        "grafana-enterprise_v1.0.1-test_333_plan9_arm-6.ext",
		"grafana-enterprise_v1.0.1-test_333_plan9_arm-6.ubuntu.docker.tar.gz": "grafana-enterprise_v1.0.1-test_333_plan9_arm-6.ext",
	}

	for k, v := range names {
		if n := pipelines.ReplaceExt(k, "ext"); n != v {
			t.Errorf("Expected '%s' without file name to equal '%s' but got '%s'", k, v, n)
		}
	}
}

func TestTarFilename(t *testing.T) {
	t.Run("It should use the correct name if Enterprise is false", func(t *testing.T) {
		distro := executil.Distribution("plan9/amd64")
		opts := pipelines.TarFileOpts{
			Edition: "",
			Version: "v1.0.1-test",
			BuildID: "333",
			Distro:  distro,
		}

		expected := "grafana_v1.0.1-test_333_plan9_amd64.tar.gz"
		if name := pipelines.TarFilename(opts); name != expected {
			t.Errorf("name '%s' does not match expected name '%s'", name, expected)
		}
	})
	t.Run("It should use the correct name if Enterprise is false", func(t *testing.T) {
		distro := executil.Distribution("plan9/amd64")
		opts := pipelines.TarFileOpts{
			Edition: "",
			Version: "v1.0.1-pre",
			BuildID: "333",
			Distro:  distro,
		}

		expected := "grafana_v1.0.1_333_plan9_amd64.tar.gz"
		if name := pipelines.TarFilename(opts); name != expected {
			t.Errorf("name '%s' does not match expected name '%s'", name, expected)
		}
	})
	t.Run("It should use the correct name if Enterprise is true", func(t *testing.T) {
		distro := executil.Distribution("plan9/amd64")
		opts := pipelines.TarFileOpts{
			Edition: "enterprise",
			Version: "v1.0.1-test",
			BuildID: "333",
			Distro:  distro,
		}

		expected := "grafana-enterprise_v1.0.1-test_333_plan9_amd64.tar.gz"
		if name := pipelines.TarFilename(opts); name != expected {
			t.Errorf("name '%s' does not match expected name '%s'", name, expected)
		}
	})
	t.Run("It should use include the arch version if one is supplied in the distro", func(t *testing.T) {
		distro := executil.Distribution("plan9/arm/v6")
		opts := pipelines.TarFileOpts{
			Edition: "enterprise",
			Version: "v1.0.1-test",
			BuildID: "333",
			Distro:  distro,
		}

		expected := "grafana-enterprise_v1.0.1-test_333_plan9_arm-6.tar.gz"
		if name := pipelines.TarFilename(opts); name != expected {
			t.Errorf("name '%s' does not match expected name '%s'", name, expected)
		}
	})
	t.Run("It should support editions with multiple hyphens", func(t *testing.T) {
		distro := executil.Distribution("plan9/arm/v6")
		opts := pipelines.TarFileOpts{
			Edition: "enterprise-rpi",
			Version: "v1.0.1-test",
			BuildID: "333",
			Distro:  distro,
		}

		expected := "grafana-enterprise-rpi_v1.0.1-test_333_plan9_arm-6.tar.gz"
		if name := pipelines.TarFilename(opts); name != expected {
			t.Errorf("name '%s' does not match expected name '%s'", name, expected)
		}
	})
}

func TestOptsFromFile(t *testing.T) {
	t.Run("It should get the correct tar file opts from a valid name", func(t *testing.T) {
		name := "grafana-enterprise_v1.0.1-test_333_plan9_arm-6.tar.gz"
		distro := executil.Distribution("plan9/arm/v6")
		expect := pipelines.TarFileOpts{
			Edition: "enterprise",
			Version: "v1.0.1-test",
			BuildID: "333",
			Distro:  distro,
		}
		got := pipelines.TarOptsFromFileName(name)
		if got.Edition != expect.Edition {
			t.Errorf("got.Edition != expect.Edition, expected '%s'", expect.Edition)
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
	t.Run("It should consider only the basename", func(t *testing.T) {
		name := "somewhere/grafana-enterprise_v1.0.1-test_333_plan9_arm-6.tar.gz"
		distro := executil.Distribution("plan9/arm/v6")
		expect := pipelines.TarFileOpts{
			Edition: "enterprise",
			Version: "v1.0.1-test",
			BuildID: "333",
			Distro:  distro,
		}
		got := pipelines.TarOptsFromFileName(name)
		if got.Edition != expect.Edition {
			t.Errorf("got.Edition != expect.Edition, expected '%s'", expect.Edition)
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
	t.Run("It should support editions with multiple hyphens", func(t *testing.T) {
		name := "somewhere/grafana-enterprise-rpi_v1.0.1-test_333_plan9_arm-6.tar.gz"
		distro := executil.Distribution("plan9/arm/v6")
		expect := pipelines.TarFileOpts{
			Edition: "enterprise-rpi",
			Version: "v1.0.1-test",
			BuildID: "333",
			Distro:  distro,
		}
		got := pipelines.TarOptsFromFileName(name)
		if got.Edition != expect.Edition {
			t.Errorf("got.Edition != expect.Edition, expected '%s', got '%s'", expect.Edition, got.Edition)
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
