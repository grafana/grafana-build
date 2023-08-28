package containers

import (
	"context"
	"testing"

	"dagger.io/dagger"
	"github.com/stretchr/testify/require"
)

func TestExitError(t *testing.T) {
	ctx := context.Background()
	dc, err := dagger.Connect(ctx)
	require.NoError(t, err)
	container := dc.Container().From("busybox").WithExec([]string{"/bin/sh", "-c", "echo hello && false"})
	_, output := ExitError(ctx, container)
	require.Error(t, output)
	require.Equal(t, "container exited with non-zero exit code\nstdout: hello\nstderr: ", output.Error())
}
