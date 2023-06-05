package main

import (
	"sort"
	"testing"
)

var dockerMapping = []m{
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_amd64.ubuntu.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise-1.2.3-ubuntu-amd64.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_amd64.ubuntu.docker.tar.gz.sha256",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise-1.2.3-ubuntu-amd64.img.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm64.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise-1.2.3-arm64.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm64.docker.tar.gz.sha256",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise-1.2.3-arm64.img.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm-7.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise-1.2.3-armv7.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm-7.docker.tar.gz.sha256",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise-1.2.3-armv7.img.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm-7.ubuntu.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise-1.2.3-ubuntu-armv7.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm-7.ubuntu.docker.tar.gz.sha256",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise-1.2.3-ubuntu-armv7.img.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_amd64.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise-1.2.3-amd64.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_amd64.docker.tar.gz.sha256",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise-1.2.3-amd64.img.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm64.ubuntu.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise-1.2.3-ubuntu-arm64.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm64.ubuntu.docker.tar.gz.sha256",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise-1.2.3-ubuntu-arm64.img.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm-7.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-oss-1.2.3-armv7.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm-7.docker.tar.gz.sha256",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-oss-1.2.3-armv7.img.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm-7.ubuntu.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-oss-1.2.3-ubuntu-armv7.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm-7.ubuntu.docker.tar.gz.sha256",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-oss-1.2.3-ubuntu-armv7.img.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_amd64.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-oss-1.2.3-amd64.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_amd64.docker.tar.gz.sha256",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-oss-1.2.3-amd64.img.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_amd64.ubuntu.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-oss-1.2.3-ubuntu-amd64.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_amd64.ubuntu.docker.tar.gz.sha256",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-oss-1.2.3-ubuntu-amd64.img.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm64.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-oss-1.2.3-arm64.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm64.docker.tar.gz.sha256",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-oss-1.2.3-arm64.img.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm64.ubuntu.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-oss-1.2.3-ubuntu-arm64.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm64.ubuntu.docker.tar.gz.sha256",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-oss-1.2.3-ubuntu-arm64.img.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-pro_v1.2.3_102_linux_amd64.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise2-1.2.3-amd64.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana-pro_v1.2.3_102_linux_arm64.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise2-1.2.3-arm64.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana-pro_v1.2.3_102_linux_arm-7.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise2-1.2.3-armv7.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana-pro_v1.2.3_102_linux_amd64.ubuntu.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise2-1.2.3-ubuntu-amd64.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana-pro_v1.2.3_102_linux_arm64.ubuntu.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise2-1.2.3-ubuntu-arm64.img",
		},
	},
	{
		input: "gs://bucket/tag/grafana-pro_v1.2.3_102_linux_arm-7.ubuntu.docker.tar.gz",
		output: []string{
			"gs://bucket/artifacts/docker/1.2.3/grafana-enterprise2-1.2.3-ubuntu-armv7.img",
		},
	},
}

func TestMoveDocker(t *testing.T) {
	bucket := "gs://bucket"
	for _, v := range dockerMapping {
		out := DockerHandler(bucket, v.input)

		if len(out) != len(v.output) {
			t.Errorf("expected %d in output but received %d\nexpected: %v\nreceived: %v", len(v.output), len(out), v.output, out)
			continue
		}
		sort.Strings(out)
		exp := v.output
		sort.Strings(exp)

		for i := range out {
			got := out[i]
			expect := exp[i]
			if expect != got {
				t.Errorf("\nExpected %s\nReceived %s", expect, got)
			}
		}
	}
}
