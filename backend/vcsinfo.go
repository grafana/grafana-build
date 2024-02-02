package backend

import (
	"fmt"
	"strings"

	"dagger.io/dagger"
)

type VCSInfo struct {
	Version          string
	Commit           *dagger.File
	EnterpriseCommit *dagger.File
	Branch           *dagger.File
}

func WithVCSInfo(c *dagger.Container, info *VCSInfo, enterprise bool) *dagger.Container {
	c = c.
		WithFile(".buildinfo.commit", info.Commit).
		WithFile(".buildinfo.branch", info.Branch)

	if enterprise {
		return c.WithFile(".buildinfo.enterprise-commit", info.EnterpriseCommit)
	}

	return c
}

// VCSInfo gets the VCS data from the directory 'src', writes them to a file on the given container, and returns the files which can be used in other containers.
func GetVCSInfo(d *dagger.Client, src *dagger.Directory, version string, enterprise bool) *VCSInfo {
	c := d.Container().From("alpine/git").
		WithEntrypoint([]string{}).
		WithMountedDirectory("/src/.git", src.Directory(".git")).
		WithWorkdir("/src").
		WithExec([]string{"/bin/sh", "-c", "git rev-parse HEAD > .buildinfo.commit"}).
		WithExec([]string{"/bin/sh", "-c", "git rev-parse --abbrev-ref HEAD > .buildinfo.branch"})

	info := &VCSInfo{
		Version: version,
		Commit:  c.File(".buildinfo.commit"),
		Branch:  c.File(".buildinfo.branch"),
	}

	if enterprise {
		info.EnterpriseCommit = c.File(".buildinfo.enterprise-commit")
	}

	return info
}

func (v *VCSInfo) X() []string {
	flags := []string{
		fmt.Sprintf("main.version=%s", strings.TrimPrefix(v.Version, "v")),
		`main.commit=$(cat ./.buildinfo.commit)`,
		`main.buildBranch=$(cat ./.buildinfo.branch)`,
	}

	if v.EnterpriseCommit != nil {
		flags = append(flags, `main.enterpriseCommit=$(cat ./.buildinfo.enterprise-commit)`)
	}

	return flags
}
