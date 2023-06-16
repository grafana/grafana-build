package main

import (
	"sort"
	"testing"
)

type m struct {
	input  string
	output []string
}

var targzMapping = []m{
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_darwin_amd64.tar.gz",
		output: []string{
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.darwin-amd64.tar.gz",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_darwin_amd64.tar.gz.sha256",
		output: []string{
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.darwin-amd64.tar.gz.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_amd64.tar.gz",
		output: []string{
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.linux-amd64.tar.gz",
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.linux-amd64-musl.tar.gz",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_amd64.tar.gz.sha256",
		output: []string{
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.linux-amd64.tar.gz.sha256",
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.linux-amd64-musl.tar.gz.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm64.tar.gz",
		output: []string{
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.linux-arm64-musl.tar.gz",
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.linux-arm64.tar.gz",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm64.tar.gz.sha256",
		output: []string{
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.linux-arm64-musl.tar.gz.sha256",
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.linux-arm64.tar.gz.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm-6.tar.gz",
		output: []string{
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.linux-armv6.tar.gz",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm-6.tar.gz.sha256",
		output: []string{
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.linux-armv6.tar.gz.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm-7.tar.gz",
		output: []string{
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.linux-armv7.tar.gz",
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.linux-armv7-musl.tar.gz",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_arm-7.tar.gz.sha256",
		output: []string{
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.linux-armv7.tar.gz.sha256",
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.linux-armv7-musl.tar.gz.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_windows_amd64.tar.gz",
		output: []string{
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.windows-amd64.tar.gz",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_windows_amd64.tar.gz.sha256",
		output: []string{
			"artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.windows-amd64.tar.gz.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm-6.tar.gz",
		output: []string{
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.linux-armv6.tar.gz",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm-6.tar.gz.sha256",
		output: []string{
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.linux-armv6.tar.gz.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm-7.tar.gz",
		output: []string{
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.linux-armv7.tar.gz",
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.linux-armv7-musl.tar.gz",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm-7.tar.gz.sha256",
		output: []string{
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.linux-armv7.tar.gz.sha256",
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.linux-armv7-musl.tar.gz.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_windows_amd64.tar.gz",
		output: []string{
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.windows-amd64.tar.gz",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_windows_amd64.tar.gz.sha256",
		output: []string{
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.windows-amd64.tar.gz.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_darwin_amd64.tar.gz",
		output: []string{
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.darwin-amd64.tar.gz",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_darwin_amd64.tar.gz.sha256",
		output: []string{
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.darwin-amd64.tar.gz.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_amd64.tar.gz",
		output: []string{
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.linux-amd64-musl.tar.gz",
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.linux-amd64.tar.gz",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_amd64.tar.gz.sha256",
		output: []string{
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.linux-amd64-musl.tar.gz.sha256",
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.linux-amd64.tar.gz.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm64.tar.gz",
		output: []string{
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.linux-arm64-musl.tar.gz",
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.linux-arm64.tar.gz",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_arm64.tar.gz.sha256",
		output: []string{
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.linux-arm64-musl.tar.gz.sha256",
			"artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.linux-arm64.tar.gz.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana-pro_v1.2.3-pre.4_102_linux_amd64.tar.gz",
		output: []string{
			"artifacts/downloads-enterprise2/v1.2.3-pre.4/enterprise2/release/grafana-enterprise2-1.2.3-pre.4.linux-amd64.tar.gz",
			"artifacts/downloads-enterprise2/v1.2.3-pre.4/enterprise2/release/grafana-enterprise2-1.2.3-pre.4.linux-amd64-musl.tar.gz",
		},
	},
	{
		input: "gs://bucket/tag/grafana-pro_v1.2.3-pre.4_102_linux_amd64.tar.gz.sha256",
		output: []string{
			"artifacts/downloads-enterprise2/v1.2.3-pre.4/enterprise2/release/grafana-enterprise2-1.2.3-pre.4.linux-amd64.tar.gz.sha256",
			"artifacts/downloads-enterprise2/v1.2.3-pre.4/enterprise2/release/grafana-enterprise2-1.2.3-pre.4.linux-amd64-musl.tar.gz.sha256",
		},
	},
}

func TestMoveTargz(t *testing.T) {
	for _, v := range targzMapping {
		out := TarGZHandler(v.input)
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
