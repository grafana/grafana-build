package executil

import "strings"

type VersionChannel string

const (
	Stable  VersionChannel = "stable"
	Preview VersionChannel = "preview"
	Nightly VersionChannel = "nightly"
	Test    VersionChannel = "test"
)

func GetVersionChannel(version string) VersionChannel {
	channel := Stable
	if n := strings.Split(version, "-"); len(n) != 1 {
		channel = Nightly
		if n[1] == string(Preview) {
			return Preview
		}
		if strings.Contains(n[1], string(Test)) {
			return Test
		}
	}
	return channel
}
