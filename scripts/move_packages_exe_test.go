package main

import (
	"sort"
	"testing"
)

var exeMapping = []m{
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_windows_amd64.exe",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.windows-amd64.exe",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_windows_amd64.exe.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/enterprise/release/grafana-enterprise-1.2.3.windows-amd64.exe.sha256",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_windows_amd64.exe",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.windows-amd64.exe",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_windows_amd64.exe.sha256",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.windows-amd64.exe.sha256",
		},
	},
}

func TestMoveEXEs(t *testing.T) {
	bucket := "gs://bucket"
	for _, v := range exeMapping {
		out := EXEHandler(bucket, v.input)

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
