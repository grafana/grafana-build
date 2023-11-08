package backend

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"dagger.io/dagger"
)

func GoLDFlags(flags map[string][]string) string {
	ldflags := strings.Builder{}
	for k, v := range flags {
		if v == nil {
			ldflags.WriteString(k + " ")
			continue
		}

		for _, value := range v {
			// For example, "-X 'main.version=v1.0.0'"
			ldflags.WriteString(fmt.Sprintf(`%s \"%s\" `, k, value))
		}
	}

	return ldflags.String()
}

// GoBuildCommand returns the arguments for go build to be used in 'WithExec'.
func GoBuildCommand(output string, ldflags map[string][]string, tags []string, main string) []string {
	args := []string{"go", "build",
		fmt.Sprintf("-ldflags=\"%s\"", GoLDFlags(ldflags)),
		fmt.Sprintf("-o=%s", output),
		"-trimpath",
		fmt.Sprintf("-tags=%s", strings.Join(tags, ",")),
		// Go is weird and paths referring to packages within a module to be prefixed with "./".
		// Otherwise, the path is assumed to be relative to $GOROOT
		"./" + main,
	}

	return args
}

func Build(
	builder *dagger.Container,
	src *dagger.Directory,
	distro Distribution,
	out string,
	opts *BuildOpts,
) *dagger.Directory {
	builder, vcsinfo := WithVCSInfo(builder, opts.Version, opts.Enterprise)
	ldflags := LDFlagsDynamic(vcsinfo)

	if opts.Static {
		ldflags = LDFlagsStatic(vcsinfo)
	}

	cmd := []string{
		"grafana",
		"grafana-server",
		"grafana-cli",
		"grafana-example-apiserver",
	}

	for _, v := range cmd {
		// Some CLI packages such as grafana-example-apiserver don't exist in earlier Grafana Versions <10.3
		// Below check skips building them as needed
		pkgPath := path.Join("pkg", "cmd", v)
		if _, err := os.Stat(pkgPath); errors.Is(err, os.ErrNotExist) {
			continue
		}
		cmd := GoBuildCommand(path.Join(out, v), ldflags, opts.Tags, pkgPath)
		builder = builder.
			WithExec([]string{"/bin/sh", "-c", strings.Join(cmd, " ")})
	}

	return builder.Directory(out)
}
