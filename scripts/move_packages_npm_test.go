package main

import (
	"sort"
	"testing"
)

var npmMapping = []m{
	{
		input: "file://dist/npm-artifacts/@grafana-data-1.2.3-pre.tgz",
		output: []string{
			"artifacts/npm/v1.2.3-pre/npm-artifacts/@grafana-data-1.2.3-pre.tgz",
		},
	},
}

func TestMoveNPM(t *testing.T) {
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
