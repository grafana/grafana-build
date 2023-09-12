package containers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dagger.io/dagger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type BuildInfo struct {
	Version   string
	Commit    string
	Branch    string
	Timestamp time.Time
}

func (b *BuildInfo) LDFlags() []string {
	return []string{
		fmt.Sprintf("main.version=%s", strings.TrimPrefix(b.Version, "v")),
		fmt.Sprintf("main.commit=%s", b.Commit),
		fmt.Sprintf("main.buildstamp=%d", b.Timestamp.Unix()),
		fmt.Sprintf("main.buildBranch=%s", b.Branch),
	}
}

const GitImage = "alpine/git:v2.36.3"

// GetBuildInfo runs a dagger pipeline using the alpine/git image to get information from the git repository.
// Because this function both creates containers and pulls data from them, it will actually run containers to get the build info for the git repository.
func GetBuildInfo(ctx context.Context, d *dagger.Client, dir *dagger.Directory, version string) (*BuildInfo, error) {
	ctx, span := otel.Tracer("grafana-build").Start(ctx, "get-buildinfo")
	defer span.End()

	container := d.Container().From(GitImage).
		WithMountedDirectory("/src", dir).
		WithWorkdir("/src")

	sha, err := revParseShort(ctx, container)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to parse commit")
		return nil, fmt.Errorf("failed to get repository's commit on HEAD: %w", err)
	}

	branch, err := revParseBranch(ctx, container)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to parse branch")
		return nil, fmt.Errorf("failed to get repository's branch on HEAD: %w", err)
	}

	timestamp := time.Now()

	result := &BuildInfo{
		Version:   version,
		Commit:    sha,
		Branch:    branch,
		Timestamp: timestamp,
	}
	span.SetAttributes(attribute.String("version", version), attribute.String("commit", sha), attribute.String("branch", branch), attribute.String("timestamp", timestamp.Format(time.RFC3339)))
	return result, nil
}

func revParseShort(ctx context.Context, container *dagger.Container) (string, error) {
	var err error
	c := container.WithExec([]string{"rev-parse", "--short", "HEAD"})

	c, err = ExitError(ctx, c)
	if err != nil {
		return "", err
	}

	stdout, err := c.Stdout(ctx)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(stdout), nil
}

func revParseBranch(ctx context.Context, container *dagger.Container) (string, error) {
	var err error
	c := container.WithExec([]string{"rev-parse", "--abbrev-ref", "HEAD"})

	c, err = ExitError(ctx, c)
	if err != nil {
		return "", err
	}

	stdout, err := c.Stdout(ctx)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(stdout), nil
}
