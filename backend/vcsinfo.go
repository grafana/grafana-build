package backend

import (
	"fmt"
	"time"

	"dagger.io/dagger"
)

type VCSInfo struct {
	Version   string
	Commit    *dagger.File
	Branch    *dagger.File
	Timestamp time.Time
}

// GetVCSInfo gets the VCS data from the directory 'src', writes them to a file on the given container, and returns the files which can be used in other containers.
func WithVCSInfo(container *dagger.Container, version string) (*dagger.Container, *VCSInfo) {
	c := container.
		WithExec([]string{"/bin/sh", "-c", "git rev-parse HEAD > .buildinfo.commit"}).
		WithExec([]string{"/bin/sh", "-c", "git rev-parse --abbrev-ref HEAD > .buildinfo.branch"})

	return c, &VCSInfo{
		Version:   version,
		Commit:    c.File(".buildinfo.commit"),
		Branch:    c.File(".buildinfo.branch"),
		Timestamp: time.Now(),
	}
}

func (v *VCSInfo) X() []string {
	return []string{
		fmt.Sprintf("main.version=%s", v.Version),
		`main.commit=$(cat ./.buildinfo.commit)`,
		`main.buildBranch=$(cat ./.buildinfo.branch)`,
		fmt.Sprintf("main.buildstamp=%d", v.Timestamp.Unix()),
	}
}
