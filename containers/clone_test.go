package containers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInjectURLCredentials(t *testing.T) {
	expected := "https://username:password@example.org/somepath?query=param"
	input := "https://example.org/somepath?query=param"
	output, err := injectURLCredentials(input, "username", "password")
	require.NoError(t, err)
	require.Equal(t, expected, output)
}
