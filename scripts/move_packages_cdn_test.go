package main

import (
	"sort"
	"testing"
)

var cdnMapping = []m{
	{
		input: "gs://bucket/tag/grafana_v1.2.3_102_linux_amd64/public",
		output: []string{
			"artifacts/static-assets/grafana-oss/1.2.3/public",
			"artifacts/static-assets/grafana/1.2.3/public",
		},
	},
	{
		input: "gs://bucket/tag/grafana-enterprise_v1.2.3_102_linux_amd64/public",
		output: []string{
			"artifacts/static-assets/grafana-enterprise/1.2.3/public",
		},
	},
	{
		input: "gs://bucket/tag/grafana-pro_v1.2.3_102_linux_amd64/public",
		output: []string{
			"artifacts/static-assets/grafana-pro/1.2.3/public",
		},
	},
}

func TestMoveCDN(t *testing.T) {
	for _, v := range cdnMapping {
		out := CDNHandler(v.input)

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
