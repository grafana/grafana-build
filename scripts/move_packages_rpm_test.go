package main

import (
	"sort"
	"testing"
)

var rpmMapping = []m{
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_amd64.rpm",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3-1.x86_64.rpm",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_amd64.rpm.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3-1.x86_64.rpm.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm-7.rpm",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3-1.armhfp.rpm",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm-7.rpm.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3-1.armhfp.rpm.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_aarch64.rpm",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3-1.aarch64.rpm",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm64.rpm.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3-1.aarch64.rpm.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_amd64.rpm",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3-1.x86_64.rpm",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_amd64.rpm.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3-1.x86_64.rpm.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm64.rpm",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3-1.aarch64.rpm",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm64.rpm.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3-1.aarch64.rpm.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm-7.rpm",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3-1.armhfp.rpm",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm-7.rpm.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3-1.armhfp.rpm.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-pro_v1.2.3-pre.4_102_linux_amd64.rpm",
		output: []string{
			"gs://bucket/artifacts/downloads-enterprise2/v1.2.3-pre.4/enterprise2/release/grafana-enterprise2-1.2.3~pre.4-1.x86_64.rpm",
		},
	},
	{
		input: "gs://bucket/tag/grafana-pro_v1.2.3-pre.4_102_linux_amd64.rpm.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads-enterprise2/v1.2.3-pre.4/enterprise2/release/grafana-enterprise2-1.2.3~pre.4-1.x86_64.rpm.sha256",
		},
	},
}

func TestMoverpm(t *testing.T) {
	bucket := "gs://bucket"
	for _, v := range rpmMapping {
		out := RPMHandler(bucket, v.input)

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
