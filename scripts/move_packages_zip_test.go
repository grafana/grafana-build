package main

import (
	"sort"
	"testing"
)

var zipMapping = []m{
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_windows_amd64.zip",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.windows-amd64.zip",
		},
	},
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_windows_amd64.zip",
		output: []string{
			"gs://bucket/artifacts/downloads/v1.2.3/oss/release/grafana-1.2.3.windows-amd64.zip",
		},
	},
}

func TestMoveZip(t *testing.T) {
	bucket := testBucket
	for _, v := range zipMapping {
		out := ZipHandler(bucket, v.input)

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
