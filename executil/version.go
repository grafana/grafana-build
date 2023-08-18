package executil

type VersionChannel string

const (
	Stable  VersionChannel = "stable"
	Preview VersionChannel = "preview"
	Nightly VersionChannel = "nightly"
)
