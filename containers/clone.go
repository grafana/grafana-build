package containers

import (
	"errors"
	"strconv"

	"dagger.io/dagger"
)

type GitCloneOptions struct {
	Ref   string
	URL   string
	Depth int
}

// CloneContainer returns the container definition that uses git clone to clone the 'git url' and checks out the ref provided at 'ref'.
func CloneContainer(d *dagger.Client, opts *GitCloneOptions) (*dagger.Container, error) {
	if opts.URL == "" {
		return nil, errors.New("URL can not be empty")
	}

	ref := "main"
	if opts.Ref != "" {
		ref = opts.Ref
	}

	cloneArgs := []string{"git", "clone"}
	if opts.Depth != 0 {
		cloneArgs = append(cloneArgs, "--depth", strconv.Itoa(opts.Depth))
	}

	cloneArgs = append(cloneArgs, opts.URL, "src")

	container := d.Container().From(GitImage).
		WithEntrypoint([]string{}).
		WithExec(cloneArgs).
		WithExec([]string{"git", "-C", "src", "fetch", "origin", ref}).
		WithExec([]string{"git", "-C", "src", "checkout", ref})

	return container, nil
}

// Clone returns the directory with the cloned repository ('url') and checked out ref ('ref').
func Clone(d *dagger.Client, url, ref string) (*dagger.Directory, error) {
	container, err := CloneContainer(d, &GitCloneOptions{
		URL:   url,
		Ref:   ref,
		Depth: 1,
	})

	if err != nil {
		return nil, err
	}

	return container.Directory("src"), nil
}
