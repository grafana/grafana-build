package containers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dagger.io/dagger"
)

type BuildInfo struct {
	Version   string
	Commit    string
	Branch    string
	Timestamp time.Time
}

const GitImage = "alpine/git:v2.36.3"

// GetBuildInfo runs a dagger pipeline using the alpine/git image to get information from the git repository.
// Because this function both creates containers and pulls data from them, it will actually run containers to get the build info for the git repository.
func GetBuildInfo(ctx context.Context, d *dagger.Client, dir *dagger.Directory, version string) (*BuildInfo, error) {
	container := d.Container().From(GitImage).
		WithMountedDirectory("/src", dir).
		WithWorkdir("/src")

	sha, err := revParseShort(ctx, container)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository's commit on HEAD: %w", err)
	}

	branch, err := revParseBranch(ctx, container)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository's branch on HEAD: %w", err)
	}

	timestamp := time.Now()

	return &BuildInfo{
		Version:   version,
		Commit:    sha,
		Branch:    branch,
		Timestamp: timestamp,
	}, nil
}

func revParseShort(ctx context.Context, container *dagger.Container) (string, error) {
	c := container.WithExec([]string{"rev-parse", "--short", "HEAD"})

	if err := ExitError(ctx, container); err != nil {
		return "", err
	}

	stdout, err := c.Stdout(ctx)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(stdout), nil
}

func revParseBranch(ctx context.Context, container *dagger.Container) (string, error) {
	c := container.WithExec([]string{"rev-parse", "--abbrev-ref", "HEAD"})

	if err := ExitError(ctx, container); err != nil {
		return "", err
	}

	stdout, err := c.Stdout(ctx)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(stdout), nil
}
