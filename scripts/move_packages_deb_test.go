package main

import (
	"sort"
	"testing"
)

var debMapping = []m{
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_amd64.deb",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana_1.2.3_amd64.deb",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_amd64.deb.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana_1.2.3_amd64.deb.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm-7.deb",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana_1.2.3_armhf.deb",
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana-rpi_1.2.3_armhf.deb",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm-7.deb.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana_1.2.3_armhf.deb.sha256",
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana-rpi_1.2.3_armhf.deb.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm64.deb",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana_1.2.3_arm64.deb",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm64.deb.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana_1.2.3_arm64.deb.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_amd64.deb",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise_1.2.3_amd64.deb",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_amd64.deb.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise_1.2.3_amd64.deb.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm64.deb",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise_1.2.3_arm64.deb",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm64.deb.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise_1.2.3_arm64.deb.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm-7.deb",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise_1.2.3_armhf.deb",
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-rpi_1.2.3_armhf.deb",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm-7.deb.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise_1.2.3_armhf.deb.sha256",
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-rpi_1.2.3_armhf.deb.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-pro_v1.2.3-pre.4_102_linux_amd64.deb",
		output: []string{
			"gs://bucket/artifacts/downloads-enterprise2/v1.2.3-pre.4/enterprise2/release/grafana-enterprise2_1.2.3~pre.4.amd64.deb",
		},
	},
	{
		input: "gs://bucket/tag/grafana-pro_v1.2.3-pre.4_102_linux_amd64.deb.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads-enterprise2/v1.2.3-pre.4/enterprise2/release/grafana-enterprise2_1.2.3~pre.4.amd64.deb.sha256",
		},
	},
}

func TestMoveDeb(t *testing.T) {
	bucket := "gs://bucket"
	for _, v := range debMapping {
		out := DebHandler(bucket, v.input)

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
