package containers

import (
	"context"
	"errors"
	"fmt"

	"dagger.io/dagger"
)

var (
	ErrorNonZero = errors.New("container exited with non-zero exit code")
)

// ExitError functionally replaces '(*container).ExitCode' in a more usable way.
// It will return an error with the container's stderr and stdout if the exit code is not zero.
func ExitError(ctx context.Context, container *dagger.Container) error {
	code, err := container.ExitCode(ctx)
	if err != nil {
		return err
	}

	if code == 0 {
		return nil
	}

	stdout, err := container.Stdout(ctx)
	if err != nil {
		return fmt.Errorf("could not retrieve container's stdout: %w: %d", ErrorNonZero, code)
	}

	stderr, err := container.Stderr(ctx)
	if err != nil {
		return fmt.Errorf("could not retrieve container's stderr: %w: %d\nstdout: %s", ErrorNonZero, code, stdout)
	}

	return fmt.Errorf("%w\nstdout: %s\nstderr: %s", ErrorNonZero, stdout, stderr)
}

func Run(ctx context.Context, containers []*dagger.Container) error {
	for _, v := range containers {
		if err := ExitError(ctx, v); err != nil {
			return err
		}
	}

	return nil
}
