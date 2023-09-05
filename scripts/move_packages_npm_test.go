package main

import (
	"sort"
	"testing"
)

var npmMapping = []m{
	{
		input: "file://dist/tag/grafana_10.2.0-pre_WxeEXjDuHTPB_linux_arm64/npm-artifacts/@grafana-data-10.2.0-pre.tgz",
		output: []string{
			"artifacts/npm/v10.2.0-pre/npm-artifacts/@grafana-data-v10.2.0-pre.tgz",
		},
	},
}

func TestMoveNPM(t *testing.T) {
	t.Setenv("DRONE_TAG", "10.2.0-pre")
	for _, v := range npmMapping {
		out := NPMHandler(v.input)

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
