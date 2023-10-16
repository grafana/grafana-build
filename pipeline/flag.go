package pipeline

// // The FlagOptionsFunc defines how this flag manipulates the artifact's options, so that when this flag is requested,
// // it can set the right values for the artifact.
// type FlagOptionsFunc func()

// A Flag is a single component of an artifact string.
// For example, in the artifact string `linux/amd64:targz:enterprise`, the flags are
// `linux/amd64`, `targz`, and `enterprise`.
type Flag struct {
	Name    string
	Options map[string]string
}
