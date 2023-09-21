package containers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"dagger.io/dagger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type BuildInfo struct {
	Version          string
	Commit           string
	EnterpriseCommit string
	Branch           string
	Timestamp        time.Time
}

func (b *BuildInfo) LDFlags() []string {
	flags := []string{
		fmt.Sprintf("main.version=%s", strings.TrimPrefix(b.Version, "v")),
		fmt.Sprintf("main.commit=%s", b.Commit),
		fmt.Sprintf("main.buildstamp=%d", b.Timestamp.Unix()),
		fmt.Sprintf("main.buildBranch=%s", b.Branch),
	}

	if b.EnterpriseCommit != "" {
		flags = append(flags, fmt.Sprintf("main.enterpriseCommit=%s", b.EnterpriseCommit))
	}
	return flags
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

	enterpriseSha, _ := enterpriseCommit(ctx, container)

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
		Version:          version,
		Commit:           sha,
		EnterpriseCommit: enterpriseSha,
		Branch:           branch,
		Timestamp:        timestamp,
	}
	span.SetAttributes(attribute.String("version", version), attribute.String("commit", sha), attribute.String("branch", branch), attribute.String("timestamp", timestamp.Format(time.RFC3339)))
	return result, nil
}

func enterpriseCommit(ctx context.Context, container *dagger.Container) (string, error) {
	var err error
	c := container.
		WithEntrypoint([]string{}).
		WithExec([]string{"/bin/sh", "-c", "cat /src/.enterprise-commit || return 0"})

	log.Println("Getting container exit error")
	c, err = ExitError(ctx, c)
	if err != nil {
		return "", nil
	}

	log.Println("Getting container stdout")
	stdout, err := c.Stdout(ctx)
	if err != nil {
		return "", nil
	}

	return strings.TrimSpace(stdout), nil
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
